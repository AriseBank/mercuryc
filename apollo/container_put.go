package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/AriseBank/apollo-controller/shared"
	"github.com/AriseBank/apollo-controller/shared/api"
	"github.com/AriseBank/apollo-controller/shared/logger"
	"github.com/AriseBank/apollo-controller/shared/osarch"

	log "gopkg.in/inconshreveable/log15.v2"
)

/*
 * Update configuration, or, if 'restore:snapshot-name' is present, restore
 * the named snapshot
 */
func containerPut(d *Daemon, r *http.Request) Response {
	// Get the container
	name := mux.Vars(r)["name"]
	c, err := containerLoadByName(d, name)
	if err != nil {
		return NotFound
	}

	// Validate the ETag
	etag := []interface{}{c.Architecture(), c.LocalConfig(), c.LocalDevices(), c.IsEphemeral(), c.Profiles()}
	err = etagCheck(r, etag)
	if err != nil {
		return PreconditionFailed(err)
	}

	configRaw := api.ContainerPut{}
	if err := json.NewDecoder(r.Body).Decode(&configRaw); err != nil {
		return BadRequest(err)
	}

	architecture, err := osarch.ArchitectureId(configRaw.Architecture)
	if err != nil {
		architecture = 0
	}

	var do = func(*operation) error { return nil }

	if configRaw.Restore == "" {
		// Update container configuration
		do = func(op *operation) error {
			args := containerArgs{
				Architecture: architecture,
				Description:  configRaw.Description,
				Config:       configRaw.Config,
				Devices:      configRaw.Devices,
				Ephemeral:    configRaw.Ephemeral,
				Profiles:     configRaw.Profiles}

			// FIXME: should set to true when not migrating
			err = c.Update(args, false)
			if err != nil {
				return err
			}

			return nil
		}
	} else {
		// Snapshot Restore
		do = func(op *operation) error {
			return containerSnapRestore(d, name, configRaw.Restore)
		}
	}

	resources := map[string][]string{}
	resources["containers"] = []string{name}

	op, err := operationCreate(operationClassTask, resources, nil, do, nil, nil)
	if err != nil {
		return InternalError(err)
	}

	return OperationResponse(op)
}

func containerSnapRestore(d *Daemon, name string, snap string) error {
	// normalize snapshot name
	if !shared.IsSnapshot(snap) {
		snap = name + shared.SnapshotDelimiter + snap
	}

	logger.Info(
		"RESTORE => Restoring snapshot",
		log.Ctx{
			"snapshot":  snap,
			"container": name})

	c, err := containerLoadByName(d, name)
	if err != nil {
		logger.Error(
			"RESTORE => loadcontainerAPOLLO() failed",
			log.Ctx{
				"container": name,
				"err":       err})
		return err
	}

	source, err := containerLoadByName(d, snap)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return fmt.Errorf("snapshot %s does not exist", snap)
		default:
			return err
		}
	}

	err = c.Restore(source)
	if err != nil {
		return err
	}

	return nil
}