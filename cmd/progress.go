//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cmd

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/schollz/progressbar/v3"
	"github.com/zalando/rds-health/internal/service"
	"github.com/zalando/rds-health/internal/types"
)

//
//

type silentbar int

func (silentbar) Describe(string) {}

func spinner[T any](bar *progressbar.ProgressBar, f func() (T, error)) (T, error) {
	ch := make(chan bool)

	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				bar.Add(1)
				time.Sleep(40 * time.Millisecond)
			}
		}
	}()

	val, err := f()

	ch <- false
	bar.Finish()

	return val, err
}

//
//

type Service interface {
	CheckHealthRegion(ctx context.Context, interval time.Duration) (*types.StatusRegion, error)
	CheckHealthNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error)
	ShowRegion(ctx context.Context) (*types.Region, error)
	ShowNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error)
}

type serviceWithSpinner struct {
	Service
	bar *progressbar.ProgressBar
}

func newServiceWithSpinner(conf aws.Config) Service {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionShowBytes(false),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSpinnerType(11),
	)

	return serviceWithSpinner{
		Service: service.New(conf, bar),
		bar:     bar,
	}
}

func (s serviceWithSpinner) CheckHealthRegion(ctx context.Context, interval time.Duration) (*types.StatusRegion, error) {
	return spinner(s.bar, func() (*types.StatusRegion, error) {
		return s.Service.CheckHealthRegion(ctx, interval)
	})
}

func (s serviceWithSpinner) CheckHealthNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error) {
	return spinner(s.bar, func() (*types.StatusNode, error) {
		return s.Service.CheckHealthNode(ctx, name, interval)
	})

}

func (s serviceWithSpinner) ShowRegion(ctx context.Context) (*types.Region, error) {
	return spinner(s.bar, func() (*types.Region, error) {
		return s.Service.ShowRegion(ctx)
	})
}

func (s serviceWithSpinner) ShowNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error) {
	return spinner(s.bar, func() (*types.StatusNode, error) {
		return s.Service.ShowNode(ctx, name, interval)
	})
}
