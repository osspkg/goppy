package formdata

import (
	"bytes"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_getBoundary(t *testing.T) {
	b := bytes.NewBufferString(`--415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3
Content-Disposition: form-data; name="name"

John
--415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3
Content-Disposition: form-data; name="age"

30
--415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3
Content-Disposition: form-data; name="private"


--415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3
Content-Disposition: form-data; name="time"

2024-01-01T12:00:00Z
--415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3--`)

	bon, err := getBoundary(b)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "415d58eaf13f82abc5dc35ac97b61f0b7671988e9c5309375ca5dd268fd3", bon)

	b = bytes.NewBufferString(`fgbsfgsfffgd
fgfdgdfsg
gdfgdfsgdf`)
	bon, err = getBoundary(b)
	casecheck.Error(t, err)
}
