package nodis

import "fmt"

type DocumentID uint64
type ObjectID uint64
type FieldID uint32
type PipelineID uint64

type objectKey struct {
	document DocumentID
	object   ObjectID
}

func validName(s string) bool {
	if len(s) == 0 || len(s) > 255 {
		return false
	}
	for i := range len(s) {
		if s[i] == 0 || s[i] == '/' {
			return false
		}
	}
	return true
}

func resourceName(document DocumentID, object ObjectID, field FieldID) string {
	return fmt.Sprintf("%d/%d/%d", document, object, field)
}
