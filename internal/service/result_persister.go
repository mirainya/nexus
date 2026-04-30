package service

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"gorm.io/gorm"
)

type ResultPersister struct{ db *gorm.DB }

func NewResultPersister(db *gorm.DB) *ResultPersister { return &ResultPersister{db: db} }

func (p *ResultPersister) Persist(pctx *pipeline.ProcessorContext, sourceID uint, tenantID uint) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		entityNameToID := make(map[string]uint)

		for _, e := range pctx.Entities {
			aliasesJSON, _ := json.Marshal(e.Aliases)
			attrsJSON, _ := json.Marshal(e.Attributes)
			evidenceJSON, _ := json.Marshal(e.Evidence)

			var existingID uint
			if e.Attributes != nil {
				if eid, ok := e.Attributes["existing_id"]; ok {
					switch v := eid.(type) {
					case float64:
						existingID = uint(v)
					case json.Number:
						n, _ := v.Int64()
						existingID = uint(n)
					}
				}
			}

			if existingID > 0 {
				tx.Model(&model.Entity{}).Where("id = ?", existingID).Updates(map[string]any{
					"attributes": attrsJSON,
					"confidence": e.Confidence,
					"evidence":   evidenceJSON,
				})
				var existing model.Entity
				if tx.First(&existing, existingID).Error == nil {
					var oldAliases, newAliases []string
					json.Unmarshal(existing.Aliases, &oldAliases)
					seen := make(map[string]bool)
					for _, a := range oldAliases {
						seen[a] = true
						newAliases = append(newAliases, a)
					}
					for _, a := range e.Aliases {
						if !seen[a] {
							newAliases = append(newAliases, a)
						}
					}
					merged, _ := json.Marshal(newAliases)
					tx.Model(&existing).Update("aliases", merged)
				}
				entityNameToID[e.Name] = existingID
			} else {
				entity := model.Entity{
					UUID:       uuid.New().String(),
					Type:       e.Type,
					Name:       e.Name,
					Aliases:    aliasesJSON,
					Attributes: attrsJSON,
					Confidence: e.Confidence,
					SourceID:   sourceID,
					Evidence:   evidenceJSON,
					TenantID:   tenantID,
				}
				if err := tx.Create(&entity).Error; err != nil {
					return err
				}
				entityNameToID[e.Name] = entity.ID

				originalJSON, _ := json.Marshal(e)
				review := model.Review{
					EntityID:     &entity.ID,
					Status:       "pending",
					OriginalData: originalJSON,
					TenantID:     tenantID,
				}
				if err := tx.Create(&review).Error; err != nil {
					return err
				}
			}
		}

		for _, r := range pctx.Relations {
			fromID, fromOK := entityNameToID[r.From]
			toID, toOK := entityNameToID[r.To]
			if !fromOK || !toOK {
				continue
			}
			metaJSON, _ := json.Marshal(r.Metadata)
			rel := model.Relation{
				UUID:         uuid.New().String(),
				FromEntityID: fromID,
				ToEntityID:   toID,
				Type:         r.Type,
				Metadata:     metaJSON,
				Confidence:   r.Confidence,
				SourceID:     sourceID,
				TenantID:     tenantID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
