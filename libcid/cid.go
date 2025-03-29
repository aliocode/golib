package libcid

import (
	"context"

	"github.com/google/uuid"
)

type ctxCIDType string

var ctxCIDKey ctxCIDType = "cid"

func GenerateCid() string {
	return uuid.NewString()
}

func InjectToCtx(ctx context.Context, cid string) context.Context {
	return context.WithValue(ctx, ctxCIDKey, cid)
}

func ExtractFromCtx(ctx context.Context) (string, bool) {
	if cid := ctx.Value(ctxCIDKey); cid != nil {
		return cid.(string), true
	}
	return "", false
}
