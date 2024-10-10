package tempolite

import (
	"context"
	"fmt"
	"testing"
)

type testMessageActivitySimple struct {
	Message string
}

type testSimpleActivity struct {
	SpecialValue string
}

func (h testSimpleActivity) Run(ctx ActivityContext, task testMessageActivitySimple) (int, string, error) {

	fmt.Println("testSimpleActivity: ", task.Message)

	return 420, "cool", nil
}

// go test -timeout 30s -v -count=1 -run ^TestActivitySimple$ .
func TestActivitySimple(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-activity-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterActivity(AsActivity[testSimpleActivity](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	if _, err := tp.EnqueueActivity(As[testSimpleActivity](), testMessageActivitySimple{"hello"}); err != nil {
		t.Fatalf("EnqueueActivity failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}
