package services

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/CPU-commits/Intranet_BNews/src/db"
	"github.com/CPU-commits/Intranet_BNews/src/forms"
	"github.com/CPU-commits/Intranet_BNews/src/res"
	"github.com/CPU-commits/Intranet_BNews/src/stack"
	"github.com/gosimple/slug"
	nats_package "github.com/nats-io/nats.go"
)

func (n *NewsService) UploadNews() {
	nats.Queue("upload_news", func(m *nats_package.Msg) {
		var data stack.NatsGolangReq

		err := json.Unmarshal(m.Data, &data)
		if err != nil {
			return
		}
		payload := make(map[string]interface{})
		v := reflect.ValueOf(data.Data)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)
				payload[key.String()] = strct.Interface()
			}
		} else {
			return
		}
		modelNews, err := newsModel.NewModel(forms.NewsDTO{
			Title:    payload["title"].(string),
			Headline: payload["headline"].(string),
			Body:     payload["body"].(string),
		}, payload["img"].(string), slug.MakeLang(payload["title"].(string), "es"), "global", "")
		if err != nil {
			return
		}
		// Upload
		_, err = newsModel.Use().InsertOne(db.Ctx, modelNews)
		if err != nil {
			return
		}
		// Notify news
		nats.PublishEncode("notify/global", &res.Notify{
			Title: payload["title"].(string),
			Link:  fmt.Sprintf("/noticias/%s", modelNews.Url),
			Img:   payload["key"].(string),
			Type:  "global",
		})
	})
}
