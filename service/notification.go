/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package service

import (
	"encoding/json"
	// "io/ioutil"
	// "net/http"
	"regexp"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
)

// NotificationHandler handles request on /service/notifications/, which listens to registry's events.
type NotificationHandler struct {
	beego.Controller
}

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d\+json`

// Post handles POST request, and records audit log or refreshes cache based on event.
func (n *NotificationHandler) Post() {
	var notification models.Notification
	//	log.Info("Notification Handler triggered!\n")
	//	log.Infof("request body in string: %s", string(n.Ctx.Input.CopyBody()))
	err := json.Unmarshal(n.Ctx.Input.CopyBody(1<<32), &notification)

	if err != nil {
		log.Errorf("error while decoding json: %v", err)
		return
	}
	var username, action, repo, project, tag_url, tag string
	var matched bool
	for _, e := range notification.Events {
		matched, err = regexp.MatchString(manifestPattern, e.Target.MediaType)
		if err != nil {
			log.Errorf("Failed to match the media type against pattern, error: %v", err)
			matched = false
		}
		if matched && strings.HasPrefix(e.Request.UserAgent, "docker") {
			username = e.Actor.Name
			action = e.Action
			repo = e.Target.Repository

			log.Error("------------- ")
			log.Errorf("---- event's actor: %v", username)
			log.Errorf("---- event's target repository is %v", e.Target.Repository)
			log.Errorf("---- event's target URL is %v", e.Target.URL)
			log.Errorf("---- event's target digest is %v", e.Target.Digest)
			log.Error("------------- ")

			tag_url = e.Target.URL
			if strings.Contains(tag_url, ":") {
				tag = tag_url[strings.LastIndex(tag_url, ":"):]
				// log.Errorf("--- tag name is %v", tag)
			}
			result, err1 := svc_utils.RegistryAPIGet(tag_url, username)

			// response, err1 := http.Get(tag_url)
			if err1 != nil {
				log.Errorf("Failed to get manifests for repo, repo name: %s, tag: %s, error: %v", repo, tag, err1)
				return
			}
			// result, err2 := ioutil.ReadAll(response.Body)
			// if err2 != nil {
			// 	log.Errorf("Failed to get manifests for repo, repo name: %s, tag: %s, error: %v", repo, tag, err2)

			// }

			// defer response.Body.Close()
			// log.Debugf("--- notification status code: %v", response.StatusCode)
			// log.Debugf("--- notification response: %v", result)

			mani := models.Manifest{}
			err = json.Unmarshal(result, &mani)
			if err != nil {
				log.Errorf("Failed to decode json from response for manifests, repo name: %s, tag: %s, error: %v", repo, tag, err)
				return
			}

			log.Debugf("---- manifest : %v", mani)
			log.Debugf("---- manifest Name : %v", mani.Name)
			log.Debugf("---- manifest Tag : %v", mani.Tag)

			if strings.Contains(repo, "/") {
				project = repo[0:strings.LastIndex(repo, "/")]
			}
			if username == "" {
				username = "anonymous"
			}
			go dao.AccessLog(username, project, repo, mani.Tag, action)
			if action == "push" {
				go func() {
					err2 := svc_utils.RefreshCatalogCache()
					if err2 != nil {
						log.Errorf("Error happens when refreshing cache: %v", err2)
					}
				}()
			}
		}
	}

}

// Render returns nil as it won't render any template.
func (n *NotificationHandler) Render() error {
	return nil
}
