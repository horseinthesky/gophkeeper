package converter

import (
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func DBSecretToPBSecret(secret db.Secret) *pb.Secret {
	return &pb.Secret{
		Owner:    secret.Owner.String,
		Kind:     secret.Kind.Int32,
		Name:     secret.Name.String,
		Value:    secret.Value,
		Created:  timestamppb.New(secret.Created.Time),
		Modified: timestamppb.New(secret.Created.Time),
		Deleted:  secret.Deleted.Bool,
	}
}

func PBSecretToDBSecret(secret *pb.Secret) db.Secret {
	return db.Secret{
		Owner: sql.NullString{
			String: secret.Owner,
			Valid: true,
		},
		Kind: sql.NullInt32{
			Int32: secret.Kind,
			Valid: true,
		},
		Name: sql.NullString{
			String: secret.Name,
			Valid: true,
		},
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
