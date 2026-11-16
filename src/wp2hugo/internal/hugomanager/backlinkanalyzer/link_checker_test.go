package backlinkanalyzer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_checkLink(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/r1", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/r2", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/r2", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	})
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	mux.HandleFunc("/chain/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.PathValue("n"))
		if n == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/chain/%d", n-1), http.StatusFound)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	result, err := checkLink(t.Context(), server.URL+"/ok")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.statusCode)
	require.False(t, result.redirected)

	result, err = checkLink(t.Context(), server.URL+"/missing")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, result.statusCode)

	result, err = checkLink(t.Context(), server.URL+"/r1")
	require.NoError(t, err)
	require.True(t, result.redirected)
	require.Equal(t, http.StatusOK, result.statusCode)
	require.Equal(t, server.URL+"/ok", result.finalURL)

	_, err = checkLink(t.Context(), server.URL+"/loop")
	require.ErrorIs(t, err, errRedirectLoop)

	// Exactly _maxRedirects redirects is allowed
	result, err = checkLink(t.Context(), fmt.Sprintf("%s/chain/%d", server.URL, _maxRedirects))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.statusCode)
	require.Equal(t, server.URL+"/chain/0", result.finalURL)

	// One more than _maxRedirects is not
	_, err = checkLink(t.Context(), fmt.Sprintf("%s/chain/%d", server.URL, _maxRedirects+1))
	require.ErrorIs(t, err, errTooManyRedirects)
}
