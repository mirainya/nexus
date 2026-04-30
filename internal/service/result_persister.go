package service

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"gorm.io/gorm"
)

const autoConfirmThreshold = 0.8

type ResultPersister struct{ db *gorm.DB }

func NewResultPersister(db *gorm.DB) *ResultPersister { return &ResultPersister{db: db} }

func (p *ResultPersister) Persist(pctx *pipeline.ProcessorContext, sourceID uint, tenantID uint) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		entityNameToID := make(map[string]uint)

		type pendingEntity struct {
			EntityID uint
			Data     pipeline.EntityData
		}
		var pendingReview []pendingEntity

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
				confirmed := e.Confidence >= autoConfirmThreshold
				entity := model.Entity{
					UUID:       uuid.New().String(),
					Type:       e.Type,
					Name:       e.Name,
					Aliases:    aliasesJSON,
					Attributes: attrsJSON,
					Confidence: e.Confidence,
					Confirmed:  confirmed,
					SourceID:   sourceID,
					Evidence:   evidenceJSON,
					TenantID:   tenantID,
				}
				if err := tx.Create(&entity).Error; err != nil {
					return err
				}
				entityNameToID[e.Name] = entity.ID

				if !confirmed {
					pendingReview = append(pendingReview, pendingEntity{
						EntityID: entity.ID,
						Data:     e,
					})
				}
			}
		}

		if len(pendingReview) > 0 {
			reviewData := make([]map[string]any, len(pendingReview))
			for i, pr := range pendingReview {
				reviewData[i] = map[string]any{
					"entity_id":  pr.EntityID,
					"type":       pr.Data.Type,
					"name":       pr.Data.Name,
					"confidence": pr.Data.Confidence,
					"attributes": pr.Data.Attributes,
					"aliases":    pr.Data.Aliases,
				}
			}
			originalJSON, _ := json.Marshal(reviewData)
			review := model.Review{
				DocumentID:   &sourceID,
				Status:       "pending",
				OriginalData: originalJSON,
				TenantID:     tenantID,
			}
			if err := tx.Create(&review).Error; err != nil {
				return err
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
