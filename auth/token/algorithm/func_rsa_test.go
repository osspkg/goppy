/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package algorithm_test

import (
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/auth/token/algorithm"
)

func TestUnit_RS256(t *testing.T) {
	alg, err := algorithm.Get(algorithm.RS256)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, alg)

	key, err := alg.GenerateKeys()
	casecheck.NoError(t, err)
	casecheck.NotNil(t, key)

	strKey, err := alg.EncodeBase64(key)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, strKey)

	fmt.Println(strKey.Private)
	fmt.Println(strKey.Public)

	strKey = &algorithm.KeyString{
		Private: "MIIEpQIBAAKCAQEAxgCVW/3iWFuKSgnUQ4j2tiS7taG3pY55cto5wykgs/OuaI5aq1KOcYQPixde0vYoMgr2T+KCsPq6E5sIfmg5ay6+/rHop6eGqcbWhv0cS8ua8RVLjQJF3aOIwRNbs/2hbXzrx17SYs8z51XRc88vKFyN7i4AH+p9RaBJxQQu1gFsMuu58gUggBwp16ceAOxreSR/JUxLkpts1gYp8Y6Q/ssBNbHeBVI8pEZ+yd5flLfLZqzt4YlkNEogiOmBe39CUWkxfe/Aml+zaBbq7zzXgdLclYSKZNF6k/KxeAcjAJK9Z58uukN6VUqU8UETYoKzYn+MvBxoayHU3/748PbD8wIDAQABAoIBAEajrKWeJSNqvSJ+8TiK46HF5yX8pP0uoEuGaXcj9CPfOwjYSKa4lFMRT05LLyxKX7rCyG9lm0SynrIh7FzUqC+CBOfu5tbxYIyvgJe2M2MjJ4r9EvAisHRLRJ3FX0EOqonaOG/vd1WDILxWlJwhfWyD4KrpxGeei8TMU6UzQdZdGOcSsqSj/I/HRhBZE0AYeexgZxSmPa+o9LFUvH7l/Sg7WZfGaNAygIsLP6OYQZvm5FYaw8tczNo19vgB7xbDc8I3WOxU+1rn21oSkMhrGftzcaX61NIqRu0vNgWBilYFX/YqfnS8CyoK0k31qEHts6gOfF7TEzLVxNLJSChCMEECgYEA5OamM7o9KBG/B+TBTq29OWeCcjLH+I/mPJ6HKZa6VN0ALE2Rw7Kjeu8UK1BW/VpK10ibif7uL+WdFfYF1N3F0wr28eYuLVpOXil+tW84kJ7BvuWavuZ/xU2/qBXDwKxj3UhjyI5kzcqJao1KFxIuFwQNlHeTl5xLeCg567cGwqECgYEA3XF7GeUnEl8752iIazRxQEV79CtF9zb4t+IqVY2un89RhJreabgLrFnlB3fSalb5qxZWN5Kr7qCcGeROwUqPcHYYboYj4RltGCCr86RCZcHdWQP2xY8f+fl6UBL2xnZ6jjphdlGYRcuKUWS4fzbf1/ehx5iCRJLIOEBbrbNnEhMCgYEAmNOWK/swUcxnavHETq+ZIvaFFZHqCX6qDdcaDx5fkcFsGChCJhLjK3TsVm7xZX4fcdU8Y8odZUU8uCjmg9T9+4Xakm9IbWdZ42x4+NIlRgl4+ed6hfKHZEggqiy94ao3ksp+NK09iFitnsJusTCmLR+7oWCk3hiwGq1g3ov4q4ECgYEAn7llyITQDOFiTQTKOUFnWquDv9eirLEa70+Tp4f9Z8nbC6HFQU3+JX/lweA8hnVbunrvD0CdAQ8Z6VcTMzp7geu8raPVp1x2owuV27QkLE+MP9OrIE6fCuhWwAEdvILi3Ung2L377oPkrdbPePr0tEsqhtRLSfjxsBlDx2N6ub8CgYEAuY3C298agX9iPepfX06fR5HQLlaz+tCtQ6AtYQuAvoF7EI2WcHLehiF3iOUwyIkR77lHwcYTsylFfE15aYJIQLk/OQdkLXa4W4dkET9M3GTGNuLAc4VzmZ86Bzx/oxHMUkv/uuhiWcwLT96whNWj+EJrYHKR32EzNVMlxqB22M8=",
		Public:  "MIIBCgKCAQEAxgCVW/3iWFuKSgnUQ4j2tiS7taG3pY55cto5wykgs/OuaI5aq1KOcYQPixde0vYoMgr2T+KCsPq6E5sIfmg5ay6+/rHop6eGqcbWhv0cS8ua8RVLjQJF3aOIwRNbs/2hbXzrx17SYs8z51XRc88vKFyN7i4AH+p9RaBJxQQu1gFsMuu58gUggBwp16ceAOxreSR/JUxLkpts1gYp8Y6Q/ssBNbHeBVI8pEZ+yd5flLfLZqzt4YlkNEogiOmBe39CUWkxfe/Aml+zaBbq7zzXgdLclYSKZNF6k/KxeAcjAJK9Z58uukN6VUqU8UETYoKzYn+MvBxoayHU3/748PbD8wIDAQAB",
	}

	key, err = alg.DecodeBase64(strKey)
	casecheck.NoError(t, err)
	casecheck.NotNil(t, key)

	msg := []byte("hello world")

	sign, err := alg.Sign(key, msg)
	casecheck.NoError(t, err)

	casecheck.NoError(t, alg.Verify(key, msg, sign))
}
