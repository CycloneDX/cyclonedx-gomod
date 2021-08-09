// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package sbom

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/require"
)

func TestCalculateFileHashes(t *testing.T) {
	t.Run("AllSupported", func(t *testing.T) {
		algos := []cdx.HashAlgorithm{
			cdx.HashAlgoMD5,
			cdx.HashAlgoSHA1,
			cdx.HashAlgoSHA256,
			cdx.HashAlgoSHA384,
			cdx.HashAlgoSHA512,
			cdx.HashAlgoSHA3_256,
			cdx.HashAlgoSHA3_512,
		}

		hashes, err := CalculateFileHashes("../../NOTICE", algos...) // TODO: use another file (create a tempfile?)
		require.NoError(t, err)
		require.Len(t, hashes, 7)
		require.Equal(t, "90b8bc82c30341e88830b0ea82f18548", hashes[0].Value)
		require.Equal(t, "8767825dace783fb1570510e21ab84ad59baa39c", hashes[1].Value)
		require.Equal(t, "02fa11d51d573ee6f4e1133cb4b5c7b8ade1eeadb951875dfc2a67c0122add65", hashes[2].Value)
		require.Equal(t, "3200f7c24a80080a7d7979aaaad480749b1fc5b07f0609749d47004c7e39265569ed17b2db5eea1f961543cc7a9627f2", hashes[3].Value)
		require.Equal(t, "afef70a115ee95c3e7d966322898909964399186b9cdd877b5d7ea12352b2b5f8b54902e674875be0fc84affe86d28fdca7893b5e7da45241f3e1e646ab0f32b", hashes[4].Value)
		require.Equal(t, "436042da3bf8a7b9bebeed1913c8e6ebf3b800aaaa1864690351754ece07caea", hashes[5].Value)
		require.Equal(t, "cb6b4798adf21d3604dbf089f410edb1d2be31d958d2c859a3bf64a7c3d8b8df29c2218a47e80e026e44ff2932771123a8e5ea9019b18bdce7a0781d4379dd9a", hashes[6].Value)
	})

	t.Run("UnsupportedAlgorithm", func(t *testing.T) {
		algos := []cdx.HashAlgorithm{
			cdx.HashAlgoBlake2b_256,
			cdx.HashAlgoBlake2b_384,
			cdx.HashAlgoBlake2b_512,
			cdx.HashAlgoBlake3,
		}

		for _, algo := range algos {
			t.Run(string(algo), func(t *testing.T) {
				_, err := CalculateFileHashes("", algo)
				require.Error(t, err)
				require.Contains(t, err.Error(), "unsupported hash algorithm")
			})
		}
	})

	t.Run("NoAlgorithms", func(t *testing.T) {
		hashes, err := CalculateFileHashes("")
		require.NoError(t, err)
		require.Empty(t, hashes)
	})
}

func TestNewProperty(t *testing.T) {
	property := NewProperty("name", "value")
	require.Equal(t, "cdx:gomod:name", property.Name)
	require.Equal(t, "value", property.Value)
}

func TestNormalizeVersion(t *testing.T) {
	t.Run("With trimPrefix", func(t *testing.T) {
		module := gomod.Module{
			Version: "v1.0.0+incompatible",
		}

		NormalizeVersion(&module, true)
		require.Equal(t, "1.0.0", module.Version)
	})

	t.Run("Without trimPrefix", func(t *testing.T) {
		module := gomod.Module{
			Version: "v1.0.0+incompatible",
		}

		NormalizeVersion(&module, false)
		require.Equal(t, "v1.0.0", module.Version)
	})
}
