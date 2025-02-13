// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package file

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDHCPv4Records(t *testing.T) {
	t.Run("valid leases", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// fill temp file with valid lease lines and some comments
		_, err = tmp.WriteString("00:11:22:33:44:55 192.0.2.100\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("11:22:33:44:55:66 192.0.2.101\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("# this is a comment\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("Subscriber-ID:\"Test subscriber 1\" 192.0.2.110\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("Subscriber-ID:\"Test subscriber \\\"2\\\"\" 192.0.2.111\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("Circuit-ID:\"circuit1\" 192.0.2.111\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("Remote-ID:\"remote1\" 192.0.2.111\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("22:33:44:55:66:77 10.10.10.50,255.255.255.0,10.10.10.1\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("22:33:44:55:66:78 10.10.10.50,255.255.255.0\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("22:33:44:55:66:79 10.10.10.50,0.0.0.0\n") // obscure netmask
		require.NoError(t, err)

		records, err := LoadDHCPv4Records(tmp.Name())
		if !assert.NoError(t, err) {
			return
		}

		if assert.Equal(t, 9, len(records)) {
			if assert.Contains(t, records, LookupMAC("00:11:22:33:44:55")) {
				assert.Equal(t, net.ParseIP("192.0.2.100"), records[LookupMAC("00:11:22:33:44:55")].ip)
				assert.Equal(t, net.IPMask(nil), records[LookupMAC("00:11:22:33:44:55")].netmask)
				assert.Equal(t, net.IP(nil), records[LookupMAC("00:11:22:33:44:55")].gateway)
			}
			if assert.Contains(t, records, LookupMAC("11:22:33:44:55:66")) {
				assert.Equal(t, net.ParseIP("192.0.2.101"), records[LookupMAC("11:22:33:44:55:66")].ip)
				assert.Equal(t, net.IPMask(nil), records[LookupMAC("11:22:33:44:55:66")].netmask)
				assert.Equal(t, net.IP(nil), records[LookupMAC("11:22:33:44:55:66")].gateway)
			}
			if assert.Contains(t, records, LookupSubscriberID("Test subscriber 1")) {
				assert.Equal(t, net.ParseIP("192.0.2.110"), records[LookupSubscriberID("Test subscriber 1")].ip)
			}
			if assert.Contains(t, records, LookupSubscriberID("Test subscriber \"2\"")) {
				assert.Equal(t, net.ParseIP("192.0.2.111"), records[LookupSubscriberID("Test subscriber \"2\"")].ip)
			}
			if assert.Contains(t, records, LookupCircuitID("circuit1")) {
				assert.Equal(t, net.ParseIP("192.0.2.111"), records[LookupCircuitID("circuit1")].ip)
			}
			if assert.Contains(t, records, LookupRemoteID("remote1")) {
				assert.Equal(t, net.ParseIP("192.0.2.111"), records[LookupRemoteID("remote1")].ip)
			}
			if assert.Contains(t, records, LookupMAC("22:33:44:55:66:77")) {
				assert.Equal(t, net.ParseIP("10.10.10.50"), records[LookupMAC("22:33:44:55:66:77")].ip)
				assert.Equal(t, net.IPv4Mask(255, 255, 255, 0), records[LookupMAC("22:33:44:55:66:77")].netmask)
				assert.Equal(t, net.ParseIP("10.10.10.1"), records[LookupMAC("22:33:44:55:66:77")].gateway)
			}
			if assert.Contains(t, records, LookupMAC("22:33:44:55:66:78")) {
				assert.Equal(t, net.ParseIP("10.10.10.50"), records[LookupMAC("22:33:44:55:66:78")].ip)
				assert.Equal(t, net.IPv4Mask(255, 255, 255, 0), records[LookupMAC("22:33:44:55:66:78")].netmask)
				assert.Equal(t, net.IP(nil), records[LookupMAC("22:33:44:55:66:78")].gateway)
			}
			if assert.Contains(t, records, LookupMAC("22:33:44:55:66:79")) {
				assert.Equal(t, net.ParseIP("10.10.10.50"), records[LookupMAC("22:33:44:55:66:79")].ip)
				assert.Equal(t, net.IPv4Mask(0, 0, 0, 0), records[LookupMAC("22:33:44:55:66:79")].netmask)
				assert.Equal(t, net.IP(nil), records[LookupMAC("22:33:44:55:66:79")].gateway)
			}
		}
	})

	t.Run("missing field", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with too few fields
		_, err = tmp.WriteString("foo\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("invalid end quote", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with missing end quote on Subscriber-ID
		_, err = tmp.WriteString("Subscriber-ID:\"Subscriber 3 192.0.2.120\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("invalid MAC", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with invalid MAC address to trigger an error
		_, err = tmp.WriteString("abcd 192.0.2.102\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("invalid IP address", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with invalid MAC address to trigger an error
		_, err = tmp.WriteString("22:33:44:55:66:77 bcde\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("syntatically correct IPv4 netmask which doesn't have contiguous bits set", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		_, err = tmp.WriteString("22:33:44:55:66:77 10.10.10.100,255.128.255.0\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("gateway specified, but missing netmask field", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		_, err = tmp.WriteString("22:33:44:55:66:77 10.10.10.100,,10.10.10.1\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("missing gateway (with comma as if it were intended)", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		_, err = tmp.WriteString("22:33:44:55:66:77 10.10.10.100,255.255.255.0,\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("MAC address and Subscriber-ID specified", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		_, err = tmp.WriteString("22:33:44:55:66:77 Subscriber-ID:\"testing\" 10.10.10.100\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("lease with IPv6 address", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with IPv6 address instead to trigger an error
		_, err = tmp.WriteString("00:11:22:33:44:55 2001:db8::10:1\n")
		require.NoError(t, err)
		_, err = LoadDHCPv4Records(tmp.Name())
		assert.Error(t, err)
	})
}

func TestLoadDHCPv6Records(t *testing.T) {
	t.Run("valid leases", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// fill temp file with valid lease lines and some comments
		_, err = tmp.WriteString("00:11:22:33:44:55 2001:db8::10:1\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("11:22:33:44:55:66 2001:db8::10:2\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("# this is a comment\n")
		require.NoError(t, err)

		records, err := LoadDHCPv6Records(tmp.Name())
		if !assert.NoError(t, err) {
			return
		}

		if assert.Equal(t, 2, len(records)) {
			if assert.Contains(t, records, LookupMAC("00:11:22:33:44:55")) {
				assert.Equal(t, net.ParseIP("2001:db8::10:1"), records[LookupMAC("00:11:22:33:44:55")].ip)
			}
			if assert.Contains(t, records, LookupMAC("11:22:33:44:55:66")) {
				assert.Equal(t, net.ParseIP("2001:db8::10:2"), records[LookupMAC("11:22:33:44:55:66")].ip)
			}
		}
	})

	t.Run("missing field", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with too few fields
		_, err = tmp.WriteString("foo\n")
		require.NoError(t, err)
		_, err = LoadDHCPv6Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("invalid MAC", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with invalid MAC address to trigger an error
		_, err = tmp.WriteString("abcd 2001:db8::10:3\n")
		require.NoError(t, err)
		_, err = LoadDHCPv6Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("invalid IP address", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with invalid MAC address to trigger an error
		_, err = tmp.WriteString("22:33:44:55:66:77 bcde\n")
		require.NoError(t, err)
		_, err = LoadDHCPv6Records(tmp.Name())
		assert.Error(t, err)
	})

	t.Run("lease with IPv4 address", func(t *testing.T) {
		// setup temp leases file
		tmp, err := ioutil.TempFile("", "test_plugin_file")
		require.NoError(t, err)
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()

		// add line with IPv4 address instead to trigger an error
		_, err = tmp.WriteString("00:11:22:33:44:55 192.0.2.100\n")
		require.NoError(t, err)
		_, err = LoadDHCPv6Records(tmp.Name())
		assert.Error(t, err)
	})
}

func TestHandler4(t *testing.T) {
	t.Run("unknown MAC", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
		}
		resp := &dhcpv4.DHCPv4{}
		assert.Nil(t, resp.ClientIPAddr)

		// if we handle this DHCP request, nothing should change since the lease is
		// unknown
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.False(t, stop)
		assert.Nil(t, result.YourIPAddr)
	})

	t.Run("known MAC", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
		}
		resp := &dhcpv4.DHCPv4{}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the MAC in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")
		StaticRecords = map[lookupValue]ipConfig{
			LookupMAC(mac): ipConfig{ip: clIPAddr},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)
		assert.Nil(t, net.IP(result.Options.Get(dhcpv4.OptionRouter)))
		assert.Nil(t, net.IPMask(result.Options.Get(dhcpv4.OptionSubnetMask)))

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})

	t.Run("known, including netmask (but no gateway)", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
		}
		resp := &dhcpv4.DHCPv4{
			Options: map[uint8][]byte{},
		}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the MAC in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")
		clNetmask := net.IPv4Mask(255, 255, 255, 0)
		StaticRecords = map[lookupValue]ipConfig{
			LookupMAC(mac): {
				ip:      clIPAddr,
				netmask: clNetmask,
			},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)
		assert.Nil(t, net.IP(result.Options.Get(dhcpv4.OptionRouter)))
		assert.Equal(t, clNetmask.String(), net.IPMask(result.Options.Get(dhcpv4.OptionSubnetMask)).String())

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})

	t.Run("known, including netmask and gateway", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
		}
		resp := &dhcpv4.DHCPv4{
			Options: map[uint8][]byte{},
		}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the MAC in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")
		clNetmask := net.IPv4Mask(255, 255, 255, 0)
		clRouter := net.ParseIP("192.0.2.1")
		StaticRecords = map[lookupValue]ipConfig{
			LookupMAC(mac): {
				ip:      clIPAddr,
				netmask: clNetmask,
				gateway: clRouter,
			},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)
		assert.Equal(t, clRouter.String(), net.IP(result.Options.Get(dhcpv4.OptionRouter)).String())
		assert.Equal(t, clNetmask.String(), net.IPMask(result.Options.Get(dhcpv4.OptionSubnetMask)).String())

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})

	/*
		extracted from sample tcpdump:

		Option: (82) Agent Information Option
			Length: 21
			Option 82 Suboption: (2) Agent Remote ID
				Length: 12
				Agent Remote ID: 020a00000affc60111000000
			Option 82 Suboption: (6) Subscriber ID
				Length: 5
				Subscriber ID: PORT1
		Option: (255) End
			Option End: 255
	*/
	testPacket1 := []byte("\x52\x15\x02\x0c\x02\x0a\x00\x00\x0a\xff\xc6\x01\x11\x00\x00\x00\x06\x05\x50\x4f\x52\x54\x31\xff")

	t.Run("known Subscriber-ID", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		relayOption := make(dhcpv4.Options)
		expectedSubscriberId := "PORT1"
		require.NoError(t, relayOption.FromBytes(testPacket1))

		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
			Options:      relayOption,
		}
		resp := &dhcpv4.DHCPv4{}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the Subscriber-ID in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")

		StaticRecords = map[lookupValue]ipConfig{
			LookupSubscriberID(expectedSubscriberId): ipConfig{ip: clIPAddr},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})

	t.Run("known Remote-ID", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		relayOption := make(dhcpv4.Options)
		expectedRemoteId := "\x02\x0a\x00\x00\x0a\xff\xc6\x01\x11\x00\x00\x00"
		require.NoError(t, relayOption.FromBytes(testPacket1))

		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
			Options:      relayOption,
		}
		resp := &dhcpv4.DHCPv4{}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the Remote-ID in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")

		StaticRecords = map[lookupValue]ipConfig{
			LookupRemoteID(expectedRemoteId): ipConfig{ip: clIPAddr},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})

	testPacket2 := []byte("\x52\x11\x01\x07\x01\x05\x4e\x65\x78\x75\x73\x02\x06\x88\xf0\x31\xa4\x46\xc1\xff")

	t.Run("Known Circuit-ID", func(t *testing.T) {
		// prepare DHCPv4 request
		mac := "00:11:22:33:44:55"
		claddr, _ := net.ParseMAC(mac)
		relayOption := make(dhcpv4.Options)
		expectedCircuitId := "Nexus"
		require.NoError(t, relayOption.FromBytes(testPacket2))

		req := &dhcpv4.DHCPv4{
			ClientHWAddr: claddr,
			Options:      relayOption,
		}
		resp := &dhcpv4.DHCPv4{}
		assert.Nil(t, resp.ClientIPAddr)

		// add lease for the Remote-ID in the lease map
		clIPAddr := net.ParseIP("192.0.2.100")

		StaticRecords = map[lookupValue]ipConfig{
			LookupCircuitID(expectedCircuitId): ipConfig{ip: clIPAddr},
		}

		// if we handle this DHCP request, the YourIPAddr field should be set
		// in the result
		result, stop := Handler4(req, resp)
		assert.Same(t, result, resp)
		assert.True(t, stop)
		assert.Equal(t, clIPAddr, result.YourIPAddr)

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})
}

func TestHandler6(t *testing.T) {
	t.Run("unknown MAC", func(t *testing.T) {
		// prepare DHCPv6 request
		mac := "11:22:33:44:55:66"
		claddr, _ := net.ParseMAC(mac)
		req, err := dhcpv6.NewSolicit(claddr)
		require.NoError(t, err)
		resp, err := dhcpv6.NewAdvertiseFromSolicit(req)
		require.NoError(t, err)
		assert.Equal(t, 0, len(resp.GetOption(dhcpv6.OptionIANA)))

		// if we handle this DHCP request, nothing should change since the lease is
		// unknown
		result, stop := Handler6(req, resp)
		assert.False(t, stop)
		assert.Equal(t, 0, len(result.GetOption(dhcpv6.OptionIANA)))
	})

	t.Run("known MAC", func(t *testing.T) {
		// prepare DHCPv6 request
		mac := "11:22:33:44:55:66"
		claddr, _ := net.ParseMAC(mac)
		req, err := dhcpv6.NewSolicit(claddr)
		require.NoError(t, err)
		resp, err := dhcpv6.NewAdvertiseFromSolicit(req)
		require.NoError(t, err)
		assert.Equal(t, 0, len(resp.GetOption(dhcpv6.OptionIANA)))

		// add lease for the MAC in the lease map
		clIPAddr := net.ParseIP("2001:db8::10:1")

		StaticRecords = map[lookupValue]ipConfig{
			LookupMAC(mac): ipConfig{ip: clIPAddr},
		}

		// if we handle this DHCP request, there should be a specific IANA option
		// set in the resulting response
		result, stop := Handler6(req, resp)
		assert.False(t, stop)
		if assert.Equal(t, 1, len(result.GetOption(dhcpv6.OptionIANA))) {
			opt := result.GetOneOption(dhcpv6.OptionIANA)
			assert.Contains(t, opt.String(), "IP=2001:db8::10:1")
		}

		// cleanup
		StaticRecords = make(map[lookupValue]ipConfig)
	})
}

func TestSetupFile(t *testing.T) {
	// too few arguments
	_, _, err := setupFile(false)
	assert.Error(t, err)

	// empty file name
	_, _, err = setupFile(false, "")
	assert.Error(t, err)

	// trigger error in LoadDHCPv*Records
	_, _, err = setupFile(false, "/foo/bar")
	assert.Error(t, err)

	_, _, err = setupFile(true, "/foo/bar")
	assert.Error(t, err)

	// setup temp leases file
	tmp, err := ioutil.TempFile("", "test_plugin_file")
	require.NoError(t, err)
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	t.Run("typical case", func(t *testing.T) {
		_, err = tmp.WriteString("00:11:22:33:44:55 2001:db8::10:1\n")
		require.NoError(t, err)
		_, err = tmp.WriteString("11:22:33:44:55:66 2001:db8::10:2\n")
		require.NoError(t, err)

		assert.Equal(t, 0, len(StaticRecords))

		// leases should show up in StaticRecords
		_, _, err = setupFile(true, tmp.Name())
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(StaticRecords))
		}
	})

	t.Run("autorefresh enabled", func(t *testing.T) {
		_, _, err = setupFile(true, tmp.Name(), autoRefreshArg)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(StaticRecords))
		}
		// we add more leases to the file
		// this should trigger an event to refresh the leases database
		// without calling setupFile again
		_, err = tmp.WriteString("22:33:44:55:66:77 2001:db8::10:3\n")
		require.NoError(t, err)
		// since the event is processed asynchronously, give it a little time
		time.Sleep(time.Millisecond * 100)
		// an additional record should show up in the database
		// but we should respect the locking first
		recLock.RLock()
		defer recLock.RUnlock()

		assert.Equal(t, 3, len(StaticRecords))
	})
}
