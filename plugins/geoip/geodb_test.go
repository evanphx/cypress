package geoip

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/cypress"
	"github.com/vektra/neko"
)

const testDb = "GeoLite2-City.mmdb"

func TestGeoIP(t *testing.T) {
	path := os.Getenv("GEOIP_DB")
	if path == "" {
		_, err := os.Stat(testDb)
		if err == nil {
			path = testDb
		} else {
			t.SkipNow()
		}
	}

	n := neko.Start(t)

	n.It("adds geoip information derived from an ip", func() {
		g, err := NewGeoDB()
		require.NoError(t, err)

		g.Path = path
		g.Field = "ip"

		err = g.Open()
		require.NoError(t, err)

		m := cypress.Log()
		m.Add("ip", "4.2.2.1")

		m2, err := g.Filter(m)
		require.NoError(t, err)

		lat, ok := m2.GetFloat("geoip.latitude")
		require.True(t, ok)

		long, ok := m2.GetFloat("geoip.longitude")
		require.True(t, ok)

		assert.Equal(t, lat, 38)
		assert.Equal(t, long, -97)
	})

	n.Meow()
}
