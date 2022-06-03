/*
 * Copyright (c) 2022, OpeningO
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"net/url"
	"testing"
)

func TestURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			url: "etcd://user:@@123@localhost:123",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uri, err := url.Parse(test.url)
			if err != nil {
				t.Errorf("Err() error = %v", err)
			}
			if uri.Scheme != "etcd" {
				t.Errorf("Scheme is not eq = %v", uri.Scheme)
			}
			if uri.User.Username() != "user" {
				t.Errorf("Username is not eq = %v", uri.User.Username())
			}
			if pwd, set := uri.User.Password(); !set || pwd != "@@123" {
				t.Errorf("Password is not eq = %v", pwd)
			}
			if uri.Host != "localhost:123" {
				t.Errorf("Host is not eq = %v", uri.Host)
			}
		})
	}
}