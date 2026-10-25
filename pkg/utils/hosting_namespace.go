package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

const (
	hostedInstanceIDDomain = "open-cluster-management.io/addon-instance/v1\x00"
	hostingNamespacePrefix = "open-cluster-management-"
	hostedInstanceIDLength = 20
	// 63 (DNS-1123 label) - len(prefix) - len("-") - instance ID.
	maxAddonTokenLength = 63 - len(hostingNamespacePrefix) - 1 - hostedInstanceIDLength
)

// HostedInstanceID returns a stable identifier for one ManagedClusterAddOn, used to name its resources on the
// hosting cluster. It is the first 20 hex characters of a SHA-256 over a versioned domain and the
// length-delimited managed cluster and addon names, so ambiguous pairs such as ("a", "bc") and ("ab", "c")
// produce different IDs.
func HostedInstanceID(managedClusterName, addonName string) string {
	h := sha256.New()
	h.Write([]byte(hostedInstanceIDDomain))
	for _, v := range []string{managedClusterName, addonName} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(v)))
		h.Write(length[:])
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))[:hostedInstanceIDLength]
}

// HostingNamespace returns the namespace on the hosting cluster for one Hosted-mode ManagedClusterAddOn:
// open-cluster-management-<addon-token>-<instance ID>. The addon token only helps humans; uniqueness comes
// from the instance ID, which always uses the full names.
func HostingNamespace(managedClusterName, addonName string) string {
	return hostingNamespacePrefix + addonToken(addonName) + "-" + HostedInstanceID(managedClusterName, addonName)
}

func addonToken(addonName string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(addonName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	token := strings.Trim(b.String(), "-")
	if len(token) > maxAddonTokenLength {
		token = strings.TrimRight(token[:maxAddonTokenLength], "-")
	}
	if token == "" {
		token = "addon"
	}
	return token
}
