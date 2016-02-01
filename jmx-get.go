package main

import (
	"os"
	"fmt"
	"strings"
	"bytes"
	"time"
	"net/http"
	"regexp"
	"io/ioutil"
	"encoding/json"
	"github.com/jessevdk/go-flags"
)

// {
//  "type" : "read",
//  "mbean" : "solr/items1:type=/replication,id=org.apache.solr.handler.ReplicationHandler",
//  "attribute":"replicationEnabled",
//  "target":{
//   "url" : "service:jmx:rmi:///jndi/rmi://10.0.252.46:10004/jmxrmi"
//  }
// }

type stProxyTarget struct {
	Url     string `json:"url"`
}

type stProxyRequest struct {
	Type      string `json:"type"`
	Target    stProxyTarget `json:"target"`
	Mbean     string `json:"mbean"`
	Attribute string `json:"attribute"`
}

type stRequest struct {
	Type      string `json:"type"`
	Mbean     string `json:"mbean"`
	Attribute string `json:"attribute"`
}


type stResponse struct {
	Value  json.RawMessage `json:"value"`
	Status int64  `json:status"`
}

type options struct {
	OptHost        string   `short:"H" long:"host" arg:"String" default:"127.0.0.1" description:"target host"`
	OptPort        string   `short:"p" long:"port" arg:"String" default:"10050" description:"target port"`
	OptPath        string   `short:"u" long:"path" arg:"String" default:"/jolokia" description:"target path"`
	OptProxy       string   `short:"P" long:"proxy-url" arg:"String" default:"" description:"jmx proxy host. eg http://ip:port/jolokia"`
	OptMbean       string   `short:"m" long:"mbean" arg:"String" required:"true" description:"mbean string"`
	OptAttribute   string   `short:"a" long:"attribute" arg:"String" required:"true" description:"mbean attr"`
}

func getValue(url string, reqBody string, subAttr []string) (string, error) {
	// fmt.Printf("=== %s == %s\n",url,reqBody)
	request, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(reqBody)),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{ Timeout: time.Duration(10 * time.Second) }
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	// fmt.Printf("=== %s\n",contents)
	resVal := stResponse{}
	err = json.Unmarshal(contents, &resVal)
	if err != nil {
		return "", err
	}
	if resVal.Status != 200 {
		return "", fmt.Errorf("status:%d",resVal.Status)
	}

	ret := resVal.Value
	for _, v := range subAttr {
		var dictVal map[string]json.RawMessage
		err = json.Unmarshal([]byte(ret), &dictVal)
		if err != nil {
			return "", err
		}
		if _, ok := dictVal[v]; ok {
			ret = dictVal[v]
		} else {
			return "", fmt.Errorf("not found")
		}
	}

	if regexp.MustCompile("^{").MatchString(string(ret)) {
		return string(ret), nil;
	}

	var retVal interface{}
	err = json.Unmarshal([]byte(ret), &retVal)
	if err != nil {
		return "", err
	}
	switch retVal.(type) {
    case string:
		return retVal.(string), nil;
	case nil:
		return "", fmt.Errorf("not found");
	default:
		return string(ret), nil
	}
}

func main() {
	os.Exit(_main())
}

func _main() (st int) {
	st = 1
	opts := options{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		return
	}

	var subAttribute []string
	splitedAttr := strings.Split(regexp.MustCompile("\\\\.").ReplaceAllString(opts.OptAttribute, "__escaped__dot__"),".")
	mainAttribute := regexp.MustCompile("__escaped__dot__").ReplaceAllString(splitedAttr[0], ".")
	if len(splitedAttr) > 1 {
		for _, v := range splitedAttr[1:] {
			subAttribute = append(subAttribute, regexp.MustCompile("__escaped__dot__").ReplaceAllString(v, "."))
		}
	}

	var Url string
	var pData []byte
	if !regexp.MustCompile("/$").MatchString(opts.OptPath) {
		opts.OptPath = opts.OptPath + "/"
	}
	if !regexp.MustCompile("^/").MatchString(opts.OptPath) {
		opts.OptPath = "/" + opts.OptPath
	}

	if opts.OptProxy != ""  {
		proxyTarget := stProxyTarget{fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%s/jmxrmi",opts.OptHost,opts.OptPort)}
		proxyRequest := stProxyRequest{Target:proxyTarget, Mbean:opts.OptMbean, Attribute:mainAttribute, Type:"read"}
		pData, _ = json.Marshal(proxyRequest)
		Url = opts.OptProxy
	} else {
		theRequest := stRequest{Mbean:opts.OptMbean, Attribute:mainAttribute, Type:"read"}
		pData, _ = json.Marshal(theRequest)
		Url = fmt.Sprintf("http://%s:%s%s",opts.OptHost,opts.OptPort,opts.OptPath)
	}

	resValue, err := getValue(Url, string(pData), subAttribute)
	if err != nil {
		fmt.Printf("Error: %s | mbean:%s attr:%s\n", err,  opts.OptMbean, opts.OptAttribute );
		return
	}
	st = 0
	fmt.Printf("%s\n",resValue)
	return
}

