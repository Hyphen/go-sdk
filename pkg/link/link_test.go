package link

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Hyphen/hyphen-go-sdk/internal/client"
	"github.com/stretchr/testify/assert"
)

// FakeHTTPClient is a fake implementation of client.HTTPClient for testing
type FakeHTTPClient struct {
	GetFake    func(ctx context.Context, url string, headers map[string]string) (*client.Response, error)
	PostFake   func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error)
	PutFake    func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error)
	PatchFake  func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error)
	DeleteFake func(ctx context.Context, url string, headers map[string]string) (*client.Response, error)
}

func (f *FakeHTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
	if f.GetFake != nil {
		return f.GetFake(ctx, url, headers)
	}
	panic("Get fake not implemented")
}

func (f *FakeHTTPClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
	if f.PostFake != nil {
		return f.PostFake(ctx, url, body, headers)
	}
	panic("Post fake not implemented")
}

func (f *FakeHTTPClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
	if f.PutFake != nil {
		return f.PutFake(ctx, url, body, headers)
	}
	panic("Put fake not implemented")
}

func (f *FakeHTTPClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
	if f.PatchFake != nil {
		return f.PatchFake(ctx, url, body, headers)
	}
	panic("Patch fake not implemented")
}

func (f *FakeHTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
	if f.DeleteFake != nil {
		return f.DeleteFake(ctx, url, headers)
	}
	panic("Delete fake not implemented")
}

func TestNew(t *testing.T) {
	t.Run("creates_a_new_link_client_with_provided_options", func(t *testing.T) {
		link, err := New(
			WithAPIKey("theApiKey"),
			WithOrganizationID("theOrgId"),
			WithURIs([]string{"https://test.com"}),
		)

		assert.NoError(t, err)
		assert.NotNil(t, link)
		assert.Equal(t, "theApiKey", link.apiKey)
		assert.Equal(t, "theOrgId", link.organizationID)
		assert.Equal(t, []string{"https://test.com"}, link.uris)
	})

	t.Run("uses_environment_variables_when_options_are_not_provided", func(t *testing.T) {
		os.Setenv("HYPHEN_API_KEY", "theEnvApiKey")
		os.Setenv("HYPHEN_ORGANIZATION_ID", "theEnvOrgId")
		t.Cleanup(func() {
			os.Unsetenv("HYPHEN_API_KEY")
			os.Unsetenv("HYPHEN_ORGANIZATION_ID")
		})

		link, err := New()

		assert.NoError(t, err)
		assert.Equal(t, "theEnvApiKey", link.apiKey)
		assert.Equal(t, "theEnvOrgId", link.organizationID)
	})

	t.Run("uses_default_uris_when_not_provided", func(t *testing.T) {
		link, err := New()

		assert.NoError(t, err)
		assert.Equal(t, defaultLinkURIs, link.uris)
	})

	t.Run("returns_an_error_when_api_key_starts_with_public_", func(t *testing.T) {
		link, err := New(WithAPIKey("public_invalid"))

		assert.Nil(t, link)
		assert.EqualError(t, err, "API key cannot start with \"public_\"")
	})

	t.Run("creates_client_when_no_options_provided", func(t *testing.T) {
		link, err := New()

		assert.NoError(t, err)
		assert.NotNil(t, link)
	})
}

func TestSetErrorHandler(t *testing.T) {
	t.Run("sets_the_error_handler", func(t *testing.T) {
		link, _ := New()
		handlerCalled := false
		handler := func(err error) {
			handlerCalled = true
		}

		link.SetErrorHandler(handler)
		link.emitError(assert.AnError)

		assert.True(t, handlerCalled)
	})
}

func TestGetURI(t *testing.T) {
	t.Run("returns_an_error_when_organization_id_is_empty", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "",
		}

		uri, err := link.getURI("", "", "")

		assert.Empty(t, uri)
		assert.EqualError(t, err, "organization ID is required")
	})

	t.Run("constructs_uri_with_organization_id_only", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
		}

		uri, err := link.getURI("", "", "")

		assert.NoError(t, err)
		assert.Equal(t, "https://api.test.com/theOrgId/codes", uri)
	})

	t.Run("constructs_uri_with_prefix1", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
		}

		uri, err := link.getURI("theCode", "", "")

		assert.NoError(t, err)
		assert.Equal(t, "https://api.test.com/theOrgId/codes/theCode", uri)
	})

	t.Run("constructs_uri_with_prefix1_and_prefix2", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
		}

		uri, err := link.getURI("theCode", "qrs", "")

		assert.NoError(t, err)
		assert.Equal(t, "https://api.test.com/theOrgId/codes/theCode/qrs", uri)
	})

	t.Run("constructs_uri_with_all_prefixes", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
		}

		uri, err := link.getURI("theCode", "qrs", "theQrId")

		assert.NoError(t, err)
		assert.Equal(t, "https://api.test.com/theOrgId/codes/theCode/qrs/theQrId", uri)
	})
}

func TestCreateShortCode(t *testing.T) {
	t.Run("creates_a_short_code_successfully", func(t *testing.T) {
		expectedResponse := &ShortCodeResponse{
			ID:      "theId",
			Code:    "theCode",
			LongURL: "https://example.com",
			Domain:  "short.link",
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusCreated,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			apiKey:         "theApiKey",
			client:         fakeClient,
		}

		result, err := link.CreateShortCode(context.Background(), "https://example.com", "short.link", nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("creates_a_short_code_with_options", func(t *testing.T) {
		var actualBody map[string]interface{}
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				actualBody = body.(map[string]interface{})
				response := &ShortCodeResponse{ID: "0"}
				responseBody, _ := json.Marshal(response)
				return &client.Response{
					StatusCode: http.StatusCreated,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			apiKey:         "theApiKey",
			client:         fakeClient,
		}
		opts := &CreateShortCodeOptions{
			Code:  "customCode",
			Title: "theTitle",
			Tags:  []string{"tag1", "tag2"},
		}

		_, err := link.CreateShortCode(context.Background(), "https://example.com", "short.link", opts)

		assert.NoError(t, err)
		assert.Equal(t, "customCode", actualBody["code"])
		assert.Equal(t, "theTitle", actualBody["title"])
		assert.Equal(t, []string{"tag1", "tag2"}, actualBody["tags"])
	})

	t.Run("returns_an_error_when_organization_id_is_not_set", func(t *testing.T) {
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "",
		}

		result, err := link.CreateShortCode(context.Background(), "https://example.com", "short.link", nil)

		assert.Nil(t, result)
		assert.EqualError(t, err, "organization ID is required")
	})

	t.Run("calls_error_handler_when_create_fails", func(t *testing.T) {
		var handledError error
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return nil, assert.AnError
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
			errorHandler:   func(err error) { handledError = err },
		}

		_, err := link.CreateShortCode(context.Background(), "https://example.com", "short.link", nil)

		assert.Error(t, err)
		assert.NotNil(t, handledError)
	})

	t.Run("returns_an_error_when_status_code_is_not_201", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "Bad Request",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.CreateShortCode(context.Background(), "https://example.com", "short.link", nil)

		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to create short code: HTTP 400: Bad Request")
	})
}

func TestGetShortCode(t *testing.T) {
	t.Run("gets_a_short_code_successfully", func(t *testing.T) {
		expectedResponse := &ShortCodeResponse{
			ID:      "theId",
			Code:    "theCode",
			LongURL: "https://example.com",
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			apiKey:         "theApiKey",
			client:         fakeClient,
		}

		result, err := link.GetShortCode(context.Background(), "theCode")

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("returns_an_error_when_status_code_is_not_200", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusNotFound,
					Status:     "Not Found",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.GetShortCode(context.Background(), "theCode")

		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to get short code: HTTP 404: Not Found")
	})
}

func TestGetShortCodes(t *testing.T) {
	t.Run("gets_short_codes_successfully", func(t *testing.T) {
		expectedResponse := &GetShortCodesResponse{
			Total:    2,
			PageNum:  1,
			PageSize: 10,
			Data: []ShortCodeResponse{
				{ID: "1", Code: "code1"},
				{ID: "2", Code: "code2"},
			},
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.GetShortCodes(context.Background(), "", nil, 0, 0)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("includes_query_parameters_in_request", func(t *testing.T) {
		var actualURL string
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				actualURL = url
				response := &GetShortCodesResponse{}
				responseBody, _ := json.Marshal(response)
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		_, err := link.GetShortCodes(context.Background(), "theTitle", []string{"tag1", "tag2"}, 2, 25)

		assert.NoError(t, err)
		assert.Contains(t, actualURL, "title=theTitle")
		assert.Contains(t, actualURL, "tags=tag1%2Ctag2")
		assert.Contains(t, actualURL, "pageNum=2")
		assert.Contains(t, actualURL, "pageSize=25")
	})
}

func TestGetTags(t *testing.T) {
	t.Run("gets_tags_successfully", func(t *testing.T) {
		expectedTags := []string{"tag1", "tag2", "tag3"}
		responseBody, _ := json.Marshal(expectedTags)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.GetTags(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedTags, result)
	})
}

func TestGetCodeStats(t *testing.T) {
	t.Run("gets_code_stats_successfully", func(t *testing.T) {
		expectedResponse := &GetCodeStatsResponse{
			Clicks: ClicksStats{
				Total:  100,
				Unique: 75,
				ByDay: []ClicksByDay{
					{Date: "2024-01-01", Total: 50, Unique: 40},
					{Date: "2024-01-02", Total: 50, Unique: 35},
				},
			},
			Referrals: []any{},
			Browsers:  []any{},
			Devices:   []any{},
			Locations: []any{},
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		result, err := link.GetCodeStats(context.Background(), "theCode", startDate, endDate)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("includes_date_range_in_query_parameters", func(t *testing.T) {
		var actualURL string
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				actualURL = url
				response := &GetCodeStatsResponse{}
				responseBody, _ := json.Marshal(response)
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		_, err := link.GetCodeStats(context.Background(), "aCode", startDate, endDate)

		assert.NoError(t, err)
		assert.Contains(t, actualURL, "startDate=2024-01-01T00%3A00%3A00Z")
		assert.Contains(t, actualURL, "endDate=2024-01-31T00%3A00%3A00Z")
	})
}

func TestUpdateShortCode(t *testing.T) {
	t.Run("updates_a_short_code_successfully", func(t *testing.T) {
		expectedResponse := &ShortCodeResponse{
			ID:      "theId",
			Code:    "theCode",
			LongURL: "https://updated.com",
			Title:   "theUpdatedTitle",
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			PatchFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}
		opts := &UpdateShortCodeOptions{
			LongURL: "https://updated.com",
			Title:   "theUpdatedTitle",
		}

		result, err := link.UpdateShortCode(context.Background(), "theCode", opts)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("returns_an_error_when_status_code_is_not_200", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			PatchFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "Bad Request",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.UpdateShortCode(context.Background(), "theCode", nil)

		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to update short code: HTTP 400: Bad Request")
	})
}

func TestDeleteShortCode(t *testing.T) {
	t.Run("deletes_a_short_code_successfully", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			DeleteFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusNoContent,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		err := link.DeleteShortCode(context.Background(), "theCode")

		assert.NoError(t, err)
	})

	t.Run("returns_an_error_when_status_code_is_not_204", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			DeleteFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusNotFound,
					Status:     "Not Found",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		err := link.DeleteShortCode(context.Background(), "theCode")

		assert.EqualError(t, err, "failed to delete short code: HTTP 404: Not Found")
	})
}

func TestCreateQRCode(t *testing.T) {
	t.Run("creates_a_qr_code_successfully", func(t *testing.T) {
		expectedResponse := &QRCodeResponse{
			ID:     "theQrId",
			Title:  "theTitle",
			QRCode: "base64data",
			QRLink: "https://qr.link",
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusCreated,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.CreateQRCode(context.Background(), "theCode", nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("returns_an_error_when_status_code_is_not_201", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			PostFake: func(ctx context.Context, url string, body interface{}, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "Bad Request",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.CreateQRCode(context.Background(), "theCode", nil)

		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to create QR code: HTTP 400: Bad Request")
	})
}

func TestGetQRCode(t *testing.T) {
	t.Run("gets_a_qr_code_successfully", func(t *testing.T) {
		expectedResponse := &QRCodeResponse{
			ID:     "theQrId",
			QRCode: "base64data",
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.GetQRCode(context.Background(), "theCode", "theQrId")

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})
}

func TestGetQRCodes(t *testing.T) {
	t.Run("gets_qr_codes_successfully", func(t *testing.T) {
		expectedResponse := &GetQRCodesResponse{
			Total:    2,
			PageNum:  1,
			PageSize: 10,
			Data: []QRCodeResponse{
				{ID: "1", QRCode: "data1"},
				{ID: "2", QRCode: "data2"},
			},
		}
		responseBody, _ := json.Marshal(expectedResponse)
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		result, err := link.GetQRCodes(context.Background(), "theCode", 0, 0)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("includes_pagination_in_query_parameters", func(t *testing.T) {
		var actualURL string
		fakeClient := &FakeHTTPClient{
			GetFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				actualURL = url
				response := &GetQRCodesResponse{}
				responseBody, _ := json.Marshal(response)
				return &client.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		_, err := link.GetQRCodes(context.Background(), "aCode", 3, 50)

		assert.NoError(t, err)
		assert.Contains(t, actualURL, "pageNum=3")
		assert.Contains(t, actualURL, "pageSize=50")
	})
}

func TestDeleteQRCode(t *testing.T) {
	t.Run("deletes_a_qr_code_successfully", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			DeleteFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusNoContent,
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		err := link.DeleteQRCode(context.Background(), "theCode", "theQrId")

		assert.NoError(t, err)
	})

	t.Run("returns_an_error_when_status_code_is_not_204", func(t *testing.T) {
		fakeClient := &FakeHTTPClient{
			DeleteFake: func(ctx context.Context, url string, headers map[string]string) (*client.Response, error) {
				return &client.Response{
					StatusCode: http.StatusNotFound,
					Status:     "Not Found",
				}, nil
			},
		}
		link := &Link{
			uris:           []string{"https://api.test.com/{organizationId}/codes/"},
			organizationID: "theOrgId",
			client:         fakeClient,
		}

		err := link.DeleteQRCode(context.Background(), "theCode", "theQrId")

		assert.EqualError(t, err, "failed to delete QR code: HTTP 404: Not Found")
	})
}
