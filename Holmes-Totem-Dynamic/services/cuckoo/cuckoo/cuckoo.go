package cuckoo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type Cuckoo struct {
	URL    	string
	APIkey	string
	MAEC	bool
	Client 	*http.Client
}

type Status struct {
	Tasks     *StatusTasks     `json:"tasks"`
	Diskspace *StatusDiskspace `json:"diskspace"`
}

type StatusTasks struct {
	Running int `json:"running"`
	Pending int `json:"pending"`
}

type StatusDiskspace struct {
	Analyses *StatusSamples `json:"analyses"`
}

type StatusSamples struct {
	Total int `json:"total"`
	Free  int `json:"free"`
	Used  int `json:"used"`
}

type TasksCreateResp struct {
	TaskID int `json:"task_id"`
}

type TasksViewResp struct {
	Message string         `json:"message"`
	Task    *TasksViewTask `json:"task"`
}

type TasksViewTask struct {
	Status string `json:"status"`
}

// dalsie su procmemory, target, network, static, dropped, debug, screenshots, strings, metadata
type TasksReport struct {
	Info       *TasksReportInfo        `json:"info"`
	Signatures []*TasksReportSignature `json;"signatures"`
	Behavior   *TasksReportBehavior    `json:"behavior"`
}

type TasksReportInfo struct {
	Started float64         `json:"started"`
	Ended   float64         `json:"ended"`
	Id      int             `json:"id"`
	Machine json.RawMessage `json:"machine"` //can be TasksReportInfoMachine OR string
}

type TasksReportInfoMachine struct {
	Name string `json:"name"`
}

type TasksReportSignature struct {
	Severity    int    `json:"severity"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type TasksReportBehavior struct {
	Processes []*TasksReportBhvPcs   `json:"processes"`
	Summary   *TasksReportBhvSummary `json:"summary"`
}

type TasksReportBhvPcs struct {
	Name      string                   `json:"process_name"`
	Id        int                      `json:"process_id"`
	ParentId  int                      `json:"parent_id"`
	FirstSeen float64                  `json:"first_seen"`
	Calls     []*TasksReportBhvPcsCall `json:"calls"`
}

type TasksReportBhvPcsCall struct {
	Category  string          `json:"category"`
	Status    int             `json:"status"`
	Return    string          `json:"return"`
	Timestamp string          `json:"timestamp"`
	ThreadId  string          `json:"thread_id"`
	Repeated  int             `json:"repeated"`
	Api       string          `json:"api"`
	Arguments json.RawMessage `json:"arguments"`
	Id        int             `json:"id"`
}

type TasksReportBhvPcsCallArg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TasksReportBhvSummary struct {
	Files   []string `json:"files"`
	Keys    []string `json:"keys"`
	Mutexes []string `json:"mutexes"`
}

type FilesView struct {
	Sample *FilesViewSample `json:"sample"`
}

type FilesViewSample struct {
	SHA1     string `json:"sha1"`
	FileType string `json:"file_type"`
	FileSize int    `json:"file_size"`
	CRC32    string `json:"crc32"`
	SSDeep   string `json:"ssdeep"`
	SHA256   string `json:"sha256"`
	SHA512   string `json:"sha512"`
	Id       int    `json:"id"`
	MD5      string `json:"md5"`
}

// Structs for MAEC report
type TasksReportMAEC struct {
	Id					string						`json:"id"`
	Type				string						`json:"type"`
	SchemaVersion   	string    					`json:"schema_version"`
	MAECObjects			[]*TasksMAECObjects			`json:"maec_objects"`
	ObservableObjects 	map[string]interface{} 	`json:"observable_objects"`
}

type TasksMAECObjects struct {
	Id                  string                      `json:"id"`
	Type                string                      `json:"type"`
	Name                string                      `json:"name"`
	Timestamp           string                      `json:"timestamp"`
	OutputObjectRefs    []string                    `json:"output_object_refs"`
	InputObjectRefs     []string                    `json:"input_object_refs"`
	InstanceObjectRefs  []string                    `json:"instance_object_refs"`
	DynamicFeatures     *TasksDynamicFeatures       `json:"dynamic_features"`
	StaticFeatures      *TasksStaticFeatures        `json:"static_features"`
	AnalysisMetadata    []*TasksAnalysisMetadata    `json:"analysis_metadata"`
	TriggeredSignatures []*TasksTriggeredSignatures `json:"triggered_signatures"`
}

type TasksDynamicFeatures struct {
	ProcessTree []*TasksProcessTree `json:"process_tree"`
}

type TasksProcessTree struct {
	OrdinalPosition     string   `json:"ordinal_position"`
	ProcessRef          string   `json:"process_ref"`
	InitiatedActionRefs []string `json:"initiated_action_refs"`
}

type TasksAnalysisMetadata struct {
	IsAutomated  string   `json:"is_automated"`
	AnalysisType string   `json:"analysis_type"`
	VmRef        string   `json:"vm_ref"`
	ToolRefs     []string `json:"tool_refs"`
	Description  string   `json:"description"`
}

type TasksTriggeredSignatures struct {
	SignatureType string `json:"signature_type"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
}

type TasksStaticFeatures struct {
	Strings []string `json:"instance_object_refs"`
}


func New(URL string, APIkey string, verifySSL bool) (*Cuckoo, error) {
	tr := &http.Transport{}
	if !verifySSL {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Cuckoo{
		URL:    URL,
		APIkey: APIkey,
		Client: &http.Client{Transport: tr},
	}, nil
}

func (c *Cuckoo) GetStatus() (*Status, error) {
	r := &Status{}
	resp, status, err := c.fastGet("/cuckoo/status", r)
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return nil, err
	}

	return r, nil
}

// submitTask submits a new task to the cuckoo api.
func (c *Cuckoo) NewTask(fileBytes []byte, fileName string, params map[string]string) (int, error) {
	// add the file to the request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return 0, err
	}
	part.Write(fileBytes)

	// add the extra payload to the request
	for key, val := range params {
		err = writer.WriteField(key, val)
		if err != nil {
			return 0, err
		}
	}

	err = writer.Close()
	if err != nil {
		return 0, err
	}

	// finalize request
	request, err := http.NewRequest("POST", c.URL+"/tasks/create/file", body)
	if err != nil {
		return 0, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization","Bearer " + c.APIkey)

	// perform request
	resp, err := c.Client.Do(request)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New(resp.Status)
	}

	// parse response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	r := &TasksCreateResp{}
	if err := json.Unmarshal(respBody, r); err != nil {
		return 0, err
	}

	return r.TaskID, nil
}

func (c *Cuckoo) TaskStatus(id int) (string, error) {
	r := &TasksViewResp{}
	resp, status, err := c.fastGet(fmt.Sprintf("/tasks/view/%d", id), r)
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return "", err
	}

	if r.Message != "" {
		return "", errors.New(r.Message)
	}

	return r.Task.Status, nil
}

func (c *Cuckoo) TaskReport(id int) (*TasksReport, error) {
	r := &TasksReport{}
	resp, status, err := c.fastGet(fmt.Sprintf("/tasks/report/%d", id), r)
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return nil, err
	}

	return r, nil
}

func (c *Cuckoo) TaskReportMAEC(id int) (*TasksReportMAEC, error) {
	r := &TasksReportMAEC{}
	resp, status, err := c.fastGet(fmt.Sprintf("/tasks/report/%d/maec", id), r) // dopytuje pridany format v cuckoo
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return nil, err
	}

	return r, nil
}

func (c *Cuckoo) GetFileInfoByMD5(md5 string) (*FilesViewSample, error) {
	r := &FilesView{}
	resp, status, err := c.fastGet("/files/view/md5/"+md5, r)
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return nil, err
	}

	return r.Sample, nil
}

func (c *Cuckoo) GetFileInfoByID(id string) (*FilesViewSample, error) {
	r := &FilesView{}
	resp, status, err := c.fastGet("/files/view/id/"+id, r)
	if err != nil || status != 200 {
		if err == nil {
			err = errors.New("no-200 ret")
		}

		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return nil, err
	}

	return r.Sample, nil
}

func (c *Cuckoo) DeleteTask(id int) error {
	resp, status, err := c.fastGet(fmt.Sprintf("/tasks/delete/%d", id), nil)
	if err != nil {
		if resp != nil {
			err = errors.New(fmt.Sprintf("%s -> [%d] %s", err.Error(), status, resp))
		}

		return err
	}

	if status != 200 {
		return errors.New(fmt.Sprintf("%d - Response code not 200", status))
	}

	return nil
}

// FastGet is a wrapper for http.Get which returns only
// the important data from the request and makes sure
// everyting is closed properly.
func (c *Cuckoo) fastGet(url string, structPointer interface{}) ([]byte, int, error) {
	request, err := http.NewRequest("GET", c.URL + url, nil)
	if err != nil {
	    return nil, 0, err
	}
	request.Header.Add("Authorization","Bearer " + c.APIkey)

	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer safeResponseClose(resp)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	if structPointer != nil {
		err = json.Unmarshal(respBody, structPointer)
	}

	return respBody, resp.StatusCode, err
}

func safeResponseClose(r *http.Response) {
	if r == nil {
		return
	}

	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()
}
