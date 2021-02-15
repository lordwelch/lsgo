package gog

import (
	"bytes"
	"encoding/json"
	"reflect"
	"time"
)

type GOGalaxy struct {
	ID                         int    `json:"id"`
	Title                      string `json:"title"`
	PurchaseLink               string `json:"purchase_link"`
	Slug                       string `json:"slug"`
	ContentSystemCompatibility struct {
		Windows bool `json:"windows"`
		Osx     bool `json:"osx"`
		Linux   bool `json:"linux"`
	} `json:"content_system_compatibility"`
	Languages map[string]string `json:"languages"`
	Links     struct {
		PurchaseLink string `json:"purchase_link"`
		ProductCard  string `json:"product_card"`
		Support      string `json:"support"`
		Forum        string `json:"forum"`
	} `json:"links"`
	InDevelopment struct {
		Active bool        `json:"active"`
		Until  interface{} `json:"until"`
	} `json:"in_development"`
	IsSecret      bool      `json:"is_secret"`
	IsInstallable bool      `json:"is_installable"`
	GameType      string    `json:"game_type"`
	IsPreOrder    bool      `json:"is_pre_order"`
	ReleaseDate   time.Time `json:"release_date"`
	Images        Images    `json:"images"`
	Dlcs          DLCs      `json:"dlcs"`
	Downloads     struct {
		Installers    []Download `json:"installers"`
		Patches       []Download `json:"patches"`
		LanguagePacks []Download `json:"language_packs"`
		BonusContent  []Download `json:"bonus_content"`
	} `json:"downloads"`
	ExpandedDlcs []Product `json:"expanded_dlcs"`
	Description  struct {
		Lead             string `json:"lead"`
		Full             string `json:"full"`
		WhatsCoolAboutIt string `json:"whats_cool_about_it"`
	} `json:"description"`
	Screenshots []struct {
		ImageID              string `json:"image_id"`
		FormatterTemplateURL string `json:"formatter_template_url"`
		FormattedImages      []struct {
			FormatterName string `json:"formatter_name"`
			ImageURL      string `json:"image_url"`
		} `json:"formatted_images"`
	} `json:"screenshots"`
	Videos []struct {
		VideoURL     string `json:"video_url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Provider     string `json:"provider"`
	} `json:"videos"`
	RelatedProducts []Product `json:"related_products"`
	Changelog       string    `json:"changelog"`
}

func (gog *GOGalaxy) UnmarshalJSON(d []byte) error {
	var err error
	type T2 GOGalaxy // create new type with same structure as T but without its method set to avoid infinite `UnmarshalJSON` call stack
	x := struct {
		T2
		ReleaseDate string `json:"release_date"`
	}{} // don't forget this, if you do and 't' already has some fields set you would lose them; see second example

	if err = json.Unmarshal(d, &x); err != nil {
		return err
	}
	*gog = GOGalaxy(x.T2)

	if gog.ReleaseDate, err = time.Parse("2006-01-02T15:04:05-0700", x.ReleaseDate); err != nil {
		return err
	}
	return nil
}

type DLCs struct {
	Products []struct {
		ID           int    `json:"id"`
		Link         string `json:"link"`
		ExpandedLink string `json:"expanded_link"`
	} `json:"products"`
	AllProductsURL         string `json:"all_products_url"`
	ExpandedAllProductsURL string `json:"expanded_all_products_url"`
}

func (DLC *DLCs) UnmarshalJSON(d []byte) error {
	type T2 GOGalaxy
	var x T2
	dec := json.NewDecoder(bytes.NewBuffer(d))
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); ok {
		if delim == '[' {
			return nil
		} else {
			return json.Unmarshal(d, &x)
		}
	} else {
		return &json.UnmarshalTypeError{
			Value:  reflect.TypeOf(t).String(),
			Type:   reflect.TypeOf(*DLC),
			Offset: 0,
			Struct: "DLCs",
			Field:  ".",
		}
	}
}

type Languages map[string]string

type Download struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Os           string `json:"os"`
	Language     string `json:"language"`
	LanguageFull string `json:"language_full"`
	Version      string `json:"version"`
	TotalSize    int64  `json:"total_size"`
	Files        []File `json:"files"`
}

func (d *Download) UnmarshalJSON(data []byte) error {
	var i json.Number
	type T2 Download // create new type with same structure as T but without its method set to avoid infinite `UnmarshalJSON` call stack
	x := struct {
		T2
		ID json.RawMessage `json:"id"`
	}{} // don't forget this, if you do and 't' already has some fields set you would lose them; see second example

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*d = Download(x.T2)
	if err := json.Unmarshal(x.ID, &i); err != nil {
		if err := json.Unmarshal(x.ID, &d.ID); err != nil {
			return err
		} else {
			return nil
		}
		return err
	}
	d.ID = i.String()
	return nil
}

type File struct {
	ID       string `json:"id"`
	Size     int    `json:"size"`
	Downlink string `json:"downlink"`
}

type Product struct {
	ID                         int    `json:"id"`
	Title                      string `json:"title"`
	PurchaseLink               string `json:"purchase_link"`
	Slug                       string `json:"slug"`
	ContentSystemCompatibility struct {
		Windows bool `json:"windows"`
		Osx     bool `json:"osx"`
		Linux   bool `json:"linux"`
	} `json:"content_system_compatibility"`
	Links struct {
		PurchaseLink string `json:"purchase_link"`
		ProductCard  string `json:"product_card"`
		Support      string `json:"support"`
		Forum        string `json:"forum"`
	} `json:"links"`
	InDevelopment struct {
		Active bool      `json:"active"`
		Until  time.Time `json:"until"`
	} `json:"in_development"`
	IsSecret      bool      `json:"is_secret"`
	IsInstallable bool      `json:"is_installable"`
	GameType      string    `json:"game_type"`
	IsPreOrder    bool      `json:"is_pre_order"`
	ReleaseDate   string    `json:"release_date"`
	Images        Images    `json:"images"`
	Languages     Languages `json:"languages,omitempty"`
}

type Images struct {
	Background          string `json:"background"`
	Logo                string `json:"logo"`
	Logo2X              string `json:"logo2x"`
	Icon                string `json:"icon"`
	SidebarIcon         string `json:"sidebarIcon"`
	SidebarIcon2X       string `json:"sidebarIcon2x"`
	MenuNotificationAv  string `json:"menuNotificationAv"`
	MenuNotificationAv2 string `json:"menuNotificationAv2"`
}
