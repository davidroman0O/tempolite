package tempolite

import (
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/featureflagversion"
)

func (tp *Tempolite[T]) getOrCreateVersion(workflowType, workflowID, changeID string, minSupported, maxSupported int) (int, error) {
	key := fmt.Sprintf("%s-%s", workflowType, changeID)

	tp.logger.Debug(tp.ctx, "Checking cache for version", "key", key)

	// Check cache first
	if cachedVersion, ok := tp.versionCache.Load(key); ok {
		version := cachedVersion.(int)
		tp.logger.Debug(tp.ctx, "Found cached version", "version", version, "key", key)
		// Update version if necessary
		if version < maxSupported {
			version = maxSupported
			tp.versionCache.Store(key, version)
			tp.logger.Debug(tp.ctx, "Updated cached version", "version", version, "key", key)
		}
		tp.logger.Debug(tp.ctx, "Returning cached version", "version", version, "key", key)
		return version, nil
	}

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating transaction when getting or creating version", "error", err)
		return 0, err
	}

	var v *ent.FeatureFlagVersion

	defer func() {
		if err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
				err = fmt.Errorf("rollback error: %v (original error: %v)", rerr, err)
			}
		}
	}()

	tp.logger.Debug(tp.ctx, "Querying feature flag version", "workflowType", workflowType, "changeID", changeID)
	v, err = tx.FeatureFlagVersion.Query().
		Where(featureflagversion.WorkflowTypeEQ(workflowType)).
		Where(featureflagversion.ChangeIDEQ(changeID)).
		Only(tp.ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			tp.logger.Error(tp.ctx, "Error querying feature flag version", "error", err)
			return 0, err
		}
		tp.logger.Debug(tp.ctx, "Creating new version", "workflowType", workflowType, "changeID", changeID, "version", minSupported)
		// Version not found, create a new one
		v, err = tx.FeatureFlagVersion.Create().
			SetWorkflowType(workflowType).
			SetWorkflowID(workflowID).
			SetChangeID(changeID).
			SetVersion(minSupported).
			Save(tp.ctx)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error creating feature flag version", "error", err)
			return 0, err
		}
		tp.logger.Debug(tp.ctx, "Created new version", "workflowType", workflowType, "changeID", changeID, "version", minSupported)
	} else {
		tp.logger.Debug(tp.ctx, "Found existing version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
		// Update the version if maxSupported is greater
		if v.Version < maxSupported {
			v.Version = maxSupported
			v, err = tx.FeatureFlagVersion.UpdateOne(v).
				SetVersion(v.Version).
				SetWorkflowID(workflowID).
				Save(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Error updating feature flag version", "error", err)
				return 0, err
			}
			tp.logger.Debug(tp.ctx, "Updated version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
		}
	}

	tp.logger.Debug(tp.ctx, "Committing transaction version", "workflowType", workflowType, "changeID", changeID, "version", v.Version)
	if err = tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "error", err)
		return 0, err
	}

	// Update cache
	tp.versionCache.Store(key, v.Version)
	tp.logger.Debug(tp.ctx, "Stored version in cache", "version", v.Version, "key", key)

	return v.Version, nil
}
