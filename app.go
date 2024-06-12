package main

import (
	"context"
	"vendinha/internal/adapter/db"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	if err := db.Connect(); err != nil {
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Erro ao conectar",
			Message: err.Error(),
		})

		runtime.Quit(ctx)
	}

	a.ctx = ctx
}
