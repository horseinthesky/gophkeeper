package converter

import (
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DBSecretToPBSecret(secret db.Secret) *pb.Secret {
	return &pb.Secret{
		Owner:    secret.Owner,
		Kind:     secret.Kind,
		Name:     secret.Name,
		Value:    secret.Value,
		Created:  timestamppb.New(secret.Created.Time),
		Modified: timestamppb.New(secret.Modified.Time),
		Deleted:  secret.Deleted.Bool,
	}
}

func PBSecretToDBSecret(secret *pb.Secret) db.Secret {
	return db.Secret{
		Owner: secret.Owner,
		Kind: secret.Kind,
		Name: secret.Name,
		Value: secret.Value,
		Created: sql.NullTime{
			Time: secret.Created.AsTime(),
			Valid: true,
		},
		Modified: sql.NullTime{
			Time: secret.Modified.AsTime(),
			Valid: true,
		},
		Deleted: sql.NullBool{
			Bool: secret.Deleted,
			Valid: true,
		},
	}
}
