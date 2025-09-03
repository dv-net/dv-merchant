package aml

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_user_keys"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/jackc/pgx/v5"
)

type UserKeyDTO struct {
	Name        models.AmlKeyType
	Description string
	Value       *string
}

type UserKeysDTO struct {
	Slug models.AMLSlug
	Keys []UserKeyDTO
}

type KeysService interface {
	GetKeys(ctx context.Context, usr *models.User, slug models.AMLSlug) (*UserKeysDTO, error)
	UpdateUserKeys(ctx context.Context, usr *models.User, dto UserKeysDTO) (*UserKeysDTO, error)
	DeleteUserKeys(ctx context.Context, usr *models.User, slug models.AMLSlug) error
}

var _ KeysService = (*Service)(nil)

func (s *Service) GetKeys(ctx context.Context, usr *models.User, slug models.AMLSlug) (*UserKeysDTO, error) {
	keys, err := s.st.AmlUserKeys().FetchAllBySlug(ctx, usr.ID, slug)
	if err != nil {
		return nil, err
	}

	return s.prepareUserKeysResult(slug, keys), nil
}

func (s *Service) DeleteUserKeys(ctx context.Context, usr *models.User, slug models.AMLSlug) error {
	return s.st.AmlUserKeys().DeleteAllUserKeysBySlug(ctx, usr.ID, slug)
}

func (s *Service) UpdateUserKeys(ctx context.Context, usr *models.User, dto UserKeysDTO) (*UserKeysDTO, error) {
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		for _, key := range dto.Keys {
			if key.Name == "" {
				s.log.Error("empty key name provided", nil, "user_id", usr.ID, "slug", dto.Slug)
				return fmt.Errorf("key name cannot be empty for user %s and slug %s", usr.ID, dto.Slug)
			}

			keyID, err := s.st.AmlUserKeys(repos.WithTx(tx)).FetchServiceKeyIDBySlugAndName(ctx, dto.Slug, key.Name)
			if err != nil {
				s.log.Error("failed to fetch key ID", err, "user_id", usr.ID, "slug", dto.Slug, "key_name", key.Name)
				return fmt.Errorf("failed to fetch key ID for slug %s and name %s: %w", dto.Slug, key.Name, err)
			}

			// Remove key if value is nil
			if key.Value == nil {
				if err = s.st.AmlUserKeys(repos.WithTx(tx)).DeleteAllUserKeysBySlugAndKeyID(ctx, usr.ID, dto.Slug, keyID); err != nil {
					s.log.Error("failed to delete key", err, "user_id", usr.ID, "slug", dto.Slug, "key_name", key.Name)
					return fmt.Errorf("failed to delete key for user %s, slug %s, name %s: %w", usr.ID, dto.Slug, key.Name, err)
				}
				continue
			}

			// Key upsert
			if _, err = s.st.AmlUserKeys(repos.WithTx(tx)).CreateOrUpdateUserKeys(ctx, usr.ID, keyID, *key.Value); err != nil {
				s.log.Error("failed to create or update key", err, "user_id", usr.ID, "slug", dto.Slug, "key_name", key.Name)
				return fmt.Errorf("failed to create or update key for user %s, slug %s, name %s: %w", usr.ID, dto.Slug, key.Name, err)
			}
		}

		providerSlug, ok := slugMapping[dto.Slug]
		if !ok {
			return fmt.Errorf("unsupported or disabled provider: %s", dto.Slug)
		}

		provider, err := s.factory.GetClient(providerSlug)
		if err != nil {
			return fmt.Errorf("failed to get provider: %w", err)
		}

		_, auth, err := s.prepareServiceDataByUser(ctx, usr, prepareParams{Slug: dto.Slug, ExternalID: "1"}, repos.WithTx(tx))
		if err != nil {
			return err
		}

		if err = provider.TestRequestWithAuth(ctx, auth); err != nil {
			return errors.New("invalid credentials")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	keys, err := s.st.AmlUserKeys().FetchAllBySlug(ctx, usr.ID, dto.Slug)
	if err != nil {
		s.log.Error("failed to fetch updated keys", err, "user_id", usr.ID, "slug", dto.Slug)
		return nil, fmt.Errorf("failed to fetch updated keys for user %s and slug %s: %w", usr.ID, dto.Slug, err)
	}

	return s.prepareUserKeysResult(dto.Slug, keys), nil
}

func (s *Service) prepareUserKeysResult(slug models.AMLSlug, keys []*repo_aml_user_keys.FetchAllBySlugRow) *UserKeysDTO {
	preparedKeys := make([]UserKeyDTO, 0, len(keys))
	for _, key := range keys {
		preparedKeys = append(preparedKeys, UserKeyDTO{
			Name:        key.Name,
			Description: key.Description,
			Value:       pgtypeutils.DecodeText(key.Value),
		})
	}

	return &UserKeysDTO{Slug: slug, Keys: preparedKeys}
}
