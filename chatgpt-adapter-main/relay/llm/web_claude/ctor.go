package web_claude

import (
	"chatgpt-adapter/core/gin/inter"
	"github.com/iocgo/sdk"
	"github.com/iocgo/sdk/env"
)

func NewAdapter(env *env.Environment) inter.Adapter {
	return &api{
		env: env,
	}
}

