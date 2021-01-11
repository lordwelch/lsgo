// +build !amd64,!arm appengine !gc noasm

package lz4block

import (
	"bytes"
	"encoding/binary"
)

func decodeBlock(dst, src []byte, window *[]byte) (ret int) {
	// window := *w
	const hasError = -2
	defer func() {
		if recover() != nil {
			ret = hasError
		}
	}()

	var si, di uint
	for {
		// Literals and match lengths (token).
		b := uint(src[si])
		si++

		// Literals.
		if lLen := b >> 4; lLen > 0 {
			switch {
			case (lLen < 0xF && si+16 < uint(len(src))) && window == nil:
				// Shortcut 1
				// if we have enough room in src and dst, and the literals length
				// is small enough (0..14) then copy all 16 bytes, even if not all
				// are part of the literals.
				copy(dst[di:], src[si:si+16])
				si += lLen
				di += lLen
				if mLen := b & 0xF; mLen < 0xF {
					// Shortcut 2
					// if the match length (4..18) fits within the literals, then copy
					// all 18 bytes, even if not all are part of the literals.
					mLen += 4
					if offset := uint(src[si]) | uint(src[si+1])<<8; mLen <= offset {
						i := di - offset
						end := i + 18
						if end > uint(len(dst)) {
							// The remaining buffer may not hold 18 bytes.
							// See https://github.com/pierrec/lz4/issues/51.
							end = uint(len(dst))
						}
						copy(dst[di:], dst[i:end])
						si += 2
						di += mLen
						continue
					}
				}
			case lLen == 0xF:
				for src[si] == 0xFF {
					lLen += 0xFF
					si++
				}
				lLen += uint(src[si])
				si++
				fallthrough
			default:
				copy(dst[di:di+lLen], src[si:si+lLen])
				if window != nil {
					(*window) = append((*window), src[si:si+lLen]...)
				}
				si += lLen
				di += lLen
			}
		}
		if si == uint(len(src)) {
			return int(di)
		} else if si > uint(len(src)) {
			return hasError
		}

		offset := uint(binary.LittleEndian.Uint16(src[si:]))
		if offset == 0 {
			return hasError
		}
		si += 2

		// Match.
		mLen := b & 0xF
		if mLen == 0xF {
			for src[si] == 0xFF {
				mLen += 0xFF
				si++
			}
			mLen += uint(src[si])
			si++
		}
		mLen += minMatch

		if window == nil {
			// Copy the match.
			expanded := dst[di-offset:]
			if mLen > offset {
				// Efficiently copy the match dst[di-offset:di] into the dst slice.
				bytesToCopy := offset * (mLen / offset)
				for n := offset; n <= bytesToCopy+offset; n *= 2 {
					copy(expanded[n:], expanded[:n])
				}
				di += bytesToCopy
				mLen -= bytesToCopy
			}
			di += uint(copy(dst[di:di+mLen], expanded[:mLen]))
		} else {
			offset -= 1
			var (
				endOffset  int  = 0
				overlapLen uint = mLen - uint(offset)
			)
			if int(offset)-int(mLen) > 0 {
				endOffset = int(offset) - int(mLen)
				overlapLen = 0
			}

			match := (*window)[len((*window))-int(offset)-1 : len((*window))-endOffset]
			matchRepeated := bytes.Repeat(match, (int(overlapLen)/(int(offset)-endOffset+1))+2)
			matchWithOverlap := matchRepeated[:mLen]

			di += uint(copy(dst[di:], matchWithOverlap))
			(*window) = append((*window), matchWithOverlap...)
		}
	}
}
