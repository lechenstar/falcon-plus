// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package funcs

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/toolkits/file"
	"github.com/toolkits/sys"
)

func UrlMetrics() (L []*model.MetricValue) {
	reportUrls := g.ReportUrls()
	sz := len(reportUrls)
	if sz == 0 {
		return
	}
	hostname, err := g.Hostname()
	if err != nil {
		hostname = "None"
	}
	for furl, param := range reportUrls {
		tags := fmt.Sprintf("url=%v,timeout=%v,method=%v,payload=%v,ref=%v,src=%v", furl, param.Timeout, param.Method, param.Payload, param.RefValue, hostname)
		if ok, _ := probeUrl(furl, param.Timeout, param.Method, param.Payload, param.RefValue); !ok {
			L = append(L, GaugeValue(g.URL_CHECK_HEALTH, 0, tags))
			continue
		}
		L = append(L, GaugeValue(g.URL_CHECK_HEALTH, 1, tags))
	}
	return
}

func probeUrl(furl string, timeout string, method string, payload string, refValue string) (bool, error) {
	outpath := fmt.Sprintf("/tmp/%s.log", GetRandomString(8))
	//bs, err := sys.CmdOutBytes("curl", "--max-filesize", "102400", "-I", "-m", timeout, "-o", "/dev/null", "-s", "-w", "%{http_code}", furl)
	var bs []byte
	var err error
	if method == "GET" {
		bs, err = sys.CmdOutBytes("curl", "--max-filesize", "102400", "-m", timeout, "-o", outpath, "-w", "%{http_code}", furl)
	} else if method == "POST" {
		bs, err = sys.CmdOutBytes("curl", "-d", payload, "--max-filesize", "102400", "-m", timeout, "-o", outpath, "-w", "%{http_code}", furl)
	}
	if err != nil {
		log.Printf("probe url [%v] failed.the err is: [%v]\n", furl, err)
		return false, err
	}
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	retdata, err := file.ReadLine(reader)
	if err != nil {
		fmt.Println("read retcode failed.err is:", err)
		return false, err
	}
	retcode := string(retdata)
	if strings.TrimSpace(retcode) != "200" {
		fmt.Printf("return code [%v] is not 200.query url is [%v]", string(retcode), furl)
		return false, err
	}
	contents, err := ioutil.ReadFile(outpath)
	if len(refValue) == 0 {
		return true, err
	}
	if err == nil {
		result := strings.Replace(string(contents), "\n", "", 1)
		if strings.Contains(result, refValue) {
			return true, nil
		} else if ok, _ := regexp.MatchString(refValue, result); ok {
			return true, nil
		}
	}

	return false, nil	
}

//生成随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

