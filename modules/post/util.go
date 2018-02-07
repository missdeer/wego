// Copyright 2013 wetalk authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package post

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lunny/log"
	"github.com/missdeer/wego/models"
	"github.com/missdeer/wego/modules/utils"
	"github.com/missdeer/wego/setting"
)

/*
func ListPostsOfCategory(cat *models.Category, posts *[]models.Post) (int64, error) {
	return models.Posts().Filter("Category", cat).RelatedSel().OrderBy("-Updated").All(posts)
}

func ListPostsOfTopic(topic *models.Topic, posts *[]models.Post) (int64, error) {
	return models.Posts().Filter("Topic", topic).RelatedSel().OrderBy("-Updated").All(posts)
}*/

var mentionRegexp = regexp.MustCompile(`\B@([\d\w-_]*)`)

func FilterMentions(user *models.User, content string) {
	matches := mentionRegexp.FindAllStringSubmatch(content, -1)
	mentions := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			mentions = append(mentions, m[1])
		}
	}
	// var users []*User
	// num, err := Users().Filter("UserName__in", mentions).Filter("Follow__User", user.Id).All(&users)
	// if err == nil && num > 0 {
	// TODO mention email to user
	// }
}

func PostBrowsersAdd(uid int64, ip string, post *models.Post) {
	var key string
	if uid == 0 {
		key = ip
	} else {
		key = utils.ToStr(uid)
	}
	key = fmt.Sprintf("PCA.%d.%s", post.Id, key)
	if setting.Cache.Get(key) != nil {
		return
	}

	if err := models.UpdatePostBrowsersById(post.Id); err != nil {
		log.Error("PostCounterAdd ", err)
	}
	setting.Cache.Put(key, true, 60)
}

func PostReplysCount(post *models.Post) {
	cnt, err := models.CountCommentsByPostId(post.Id)
	if err == nil {
		post.Replys = int(cnt)
		//disable post editable
		post.CanEdit = false
		err = models.UpdateById(post.Id, post, "replys", "can_edit")
	}
	if err != nil {
		log.Error("PostReplysCount ", err)
	}
}

func FilterCommentMentions(fromUser *models.User, post *models.Post, comment *models.Comment) {
	var uri = fmt.Sprintf("post/%d", post.Id)
	var lang = setting.DefaultLang
	if fromUser.Id != post.UserId {
		var notification = models.Notification{
			FromUserId:   fromUser.Id,
			ToUserId:     post.UserId,
			Action:       setting.NOTICE_TYPE_COMMENT,
			Title:        post.Title,
			TargetId:     post.Id,
			Uri:          uri,
			Lang:         lang,
			Floor:        comment.Floor,
			Content:      comment.Message,
			ContentCache: comment.MessageCache,
			Status:       setting.NOTICE_UNREAD,
		}
		if err := models.InsertNotification(&notification); err == nil {
			//pass
		}
	}

	//check comment @
	var pattern = "[ ]*@[a-zA-Z0-9]+[ ]*"
	r := regexp.MustCompile(pattern)
	userNames := r.FindAllString(comment.Message, -1)
	for _, userName := range userNames {
		bUserName := strings.TrimPrefix(strings.TrimSpace(userName), "@")

		if user, err := models.GetUserByName(bUserName); err == nil {
			if user.Id != 0 && user.Id != post.UserId {
				notification := models.Notification{
					FromUserId:   fromUser.Id,
					ToUserId:     user.Id,
					Action:       setting.NOTICE_TYPE_COMMENT,
					Title:        post.Title,
					TargetId:     post.Id,
					Uri:          uri,
					Lang:         lang,
					Floor:        comment.Floor,
					Content:      comment.Message,
					ContentCache: comment.MessageCache,
					Status:       setting.NOTICE_UNREAD,
				}
				if err := models.InsertNotification(&notification); err == nil {
					//pass
				}
			}
		}
	}
}
