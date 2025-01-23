package data

import (
	"context"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/ent"
)

type Data struct {
	context.Context
	config *DataConfig
	comfy  *comfylite3.ComfyDB
	client *ent.Client
}

type DataConfig struct {
	memory   bool
	filePath string
}

type DataOption func(*DataConfig) error

func WithMemory() DataOption {
	return func(c *DataConfig) error {
		c.memory = true
		return nil
	}
}

func WithFilePath(filePath string) DataOption {
	return func(c *DataConfig) error {
		c.filePath = filePath
		return nil
	}
}

func New(ctx context.Context, opt ...DataOption) (*Data, error) {

	config := &DataConfig{}
	for _, o := range opt {
		if err := o(config); err != nil {
			return nil, err
		}
	}
	var err error
	var comfy *comfylite3.ComfyDB

	var comfyOptions []comfylite3.ComfyOption = []comfylite3.ComfyOption{}
	if config.memory {
		comfyOptions = append(comfyOptions, comfylite3.WithMemory())
	} else {
		comfyOptions = append(comfyOptions, comfylite3.WithPath(config.filePath))
		// Ensure folder exists
		if err := os.MkdirAll(filepath.Dir(config.filePath), 0755); err != nil {
			return nil, err
		}
	}

	// Create a new ComfyDB instance
	if comfy, err = comfylite3.New(
		comfyOptions...,
	); err != nil {
		return nil, err
	}

	// Use the OpenDB function to create a sql.DB instance with SQLite options
	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)

	// Create a new ent client
	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	// Run the auto migration tool
	if err = client.Schema.Create(ctx); err != nil {
		return nil, err
	}

	return &Data{
		Context: ctx,
		config:  config,
		client:  client,
		comfy:   comfy,
	}, nil
}

func (d *Data) Close() {
	d.client.Close()
	d.comfy.Close()
}

func (d *Data) Test() {

	// d.client.WorkflowEntity.Create().SetID(1).SetHandlerName("test").SetStatus("PENDING").SaveX(d.Context)

}
