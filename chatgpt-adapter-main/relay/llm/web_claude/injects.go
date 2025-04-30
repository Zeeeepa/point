package web_claude

import (
	"chatgpt-adapter/core/gin/inter"
	"github.com/iocgo/sdk"
)

func Injects(container *sdk.Container) error {
	return container.Provide(func(adapter inter.Adapter) inter.Adapter {
		return adapter
	}, sdk.WithName("web_claude.adapter"), sdk.WithConstructor(NewAdapter))
}

