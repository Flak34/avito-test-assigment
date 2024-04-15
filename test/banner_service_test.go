//go:build integration

package test

import (
	"avito-test-assigment/internal/model"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func Test_GetByFeatureAndTag(t *testing.T) {
	t.Run("fail", func(t *testing.T) {

		t.Run("unauthorized", func(t *testing.T) {
			client := http.Client{}
			req, err := http.NewRequest(
				"GET", "http://localhost:9001/user_banner", nil,
			)
			req.Header.Add("token", "")

			resp, err := client.Do(req)

			require.NoError(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

		t.Run("incorrect data", func(t *testing.T) {
			client := http.Client{}
			req, err := http.NewRequest(
				"GET", "http://localhost:9001/user_banner", nil,
			)
			req.Header.Add("token", token)

			resp, err := client.Do(req)

			require.NoError(t, err)
			assert.Equal(t, 400, resp.StatusCode)
		})

		t.Run("banner is not present", func(t *testing.T) {
			client := http.Client{}
			req, err := http.NewRequest(
				"GET", "http://localhost:9001/user_banner?tag_id=-88&feature_id=-845", nil,
			)
			req.Header.Add("token", token)

			resp, err := client.Do(req)

			require.NoError(t, err)
			assert.Equal(t, 404, resp.StatusCode)
		})

	})

	t.Run("success", func(t *testing.T) {
		client := http.Client{}
		req, err := http.NewRequest(
			"GET", "http://localhost:9001/user_banner", nil,
		)
		q := req.URL.Query()
		q.Add("tag_id", "1")
		q.Add("feature_id", "1")
		req.URL.RawQuery = q.Encode()

		req.Header.Add("token", token)
		require.NoError(t, err)
		tagID := new(int64)
		*tagID = 1
		c := json.RawMessage("{\"name\":\"some_name\"}")
		id, err := service.Create(context.TODO(), model.Banner{
			Content:   &c,
			IsActive:  true,
			TagIDs:    []*int64{tagID},
			FeatureID: 1,
		})
		require.NoError(t, err)

		resp, err := client.Do(req)
		err2 := bannerRepository.DeleteByID(context.Background(), id)
		require.NoError(t, err2)

		require.NoError(t, err)
		require.NoError(t, err2)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
