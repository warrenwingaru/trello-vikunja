package vikunja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
	"wingaru.me/trello-migrate/internal/models"
)

type Client struct {
	Client   *http.Client
	Logger   logger
	BaseURL  string
	Key      string
	throttle *rate.Limiter
	ctx      context.Context
}

type logger interface {
	Debugf(string, ...interface{})
}

func NewClient(key string, instanceUrl string) *Client {
	limit := rate.Every(time.Second / 8)
	return &Client{
		Key:      key,
		Client:   http.DefaultClient,
		BaseURL:  instanceUrl,
		throttle: rate.NewLimiter(limit, 1),
		ctx:      context.Background(),
	}
}

func (c *Client) WithContext(ctx context.Context) *Client {
	newC := *c
	newC.ctx = ctx
	return &newC
}

func (c *Client) Throttle() {
	c.throttle.Wait(c.ctx)

}

func (c *Client) log(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Debugf(format, args...)
	}

}

type httpClientError struct {
	msg  string
	code int
}

func (h httpClientError) Error() string {
	return h.msg
}

func (c *Client) do(req *http.Request, url string, target interface{}) error {
	resp, err := c.Client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "http request failed on %s", url)
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		msg := fmt.Sprintf("HTTP request failure on %s:\n%d: %s", url, resp.StatusCode, string(body))

		return &httpClientError{
			msg:  msg,
			code: resp.StatusCode,
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "http read error on response for %s", url)
	}
	err = json.Unmarshal(b, target)
	if err != nil {
		return nil
	}
	return nil
}

func (c *Client) put(path string, body io.Reader, target interface{}) error {
	c.Throttle()

	c.log("[vikunja] PUT %s", path)
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	req, err := http.NewRequest("PUT", url, body)

	if err != nil {
		return errors.Wrapf(err, "Invalid PUT request %s", url)
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, url, target)
}

func (c *Client) putMultipart(path string, body io.Reader, target interface{}, writer *multipart.Writer) error {
	c.Throttle()

	c.log("[vikunja] PUT %s", path)
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	req, err := http.NewRequest("PUT", url, body)

	if err != nil {
		return errors.Wrapf(err, "Invalid PUT request %s", url)
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(req, url, target)
}

func (c *Client) post(path string, body io.Reader, target interface{}) error {
	c.Throttle()

	c.log("[vikunja] POST %s", path)
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)

	req, err := http.NewRequest("POST", url, body)

	if err != nil {
		return errors.Wrapf(err, "Invalid POST request %s", url)
	}

	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, url, target)
}

func (c *Client) CreateBucket(bucket *models.Bucket) error {
	path := fmt.Sprintf("projects/%d/views/%d/buckets", bucket.ProjectID, bucket.ProjectViewID)
	data, err := json.Marshal(bucket)
	if err != nil {
		return err
	}
	err = c.put(path, bytes.NewBuffer(data), &bucket)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateBucket(bucket *models.Bucket) error {
	url := fmt.Sprintf("projects/%d/views/%d/buckets/%d", bucket.ProjectID, bucket.ProjectViewID, bucket.ID)
	data, err := json.Marshal(bucket)
	if err != nil {
		return err
	}
	err = c.post(url, bytes.NewBuffer(data), &bucket)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddTask(task *models.Task) error {

	url := fmt.Sprintf("projects/%d/tasks", task.ProjectID)
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	err = c.put(url, bytes.NewBuffer(data), &task)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddTaskComment(comment *models.TaskComment) error {

	url := fmt.Sprintf("tasks/%d/comments", comment.TaskID)
	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}
	err = c.put(url, bytes.NewBuffer(data), &comment)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddTaskAttachments(taskID int64, attachment *models.TaskAttachment) error {

	path := fmt.Sprintf("tasks/%d/attachments", taskID)

	// multipart writer
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("files", attachment.File.Name)
	c.log("[vikunja] AddTaskAttachments %s", attachment.File.Name)
	if err != nil {
		return err
	}
	_, err = part.Write(attachment.File.FileContent)
	if err != nil {
		return err
	}
	writer.Close()

	err = c.putMultipart(path, &requestBody, &attachment, writer)
	if err != nil {
		return err
	}

	return nil
}
func addBytesToRequest(buf []byte, fileName string, writer *multipart.Writer) {
	// Create a form file field for the file
	fileField, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		panic(err)
	}

	// Copy the file contents from bytes.Buffer to the form file field
	_, err = fileField.Write(buf)
	if err != nil {
		panic(err)
	}
}
