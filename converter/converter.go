package converter

import (
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func SecretToPB(secret db.Secret) *pb.Secret {
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

func PBtoSecret(pbSecret *pb.Secret) db.Secret {
	return db.Secret{
		Owner: sql.NullString{
			String: pbSecret.Owner,
			Valid: true,
		},
		Kind: sql.NullInt32{
			Int32: pbSecret.Kind,
			Valid: true,
		},
		Name: sql.NullString{
			String: pbSecret.Name,
			Valid: true,
		},
		Value: pbSecret.Value,
		Created: sql.NullTime{
			Time: pbSecret.Created.AsTime(),
			Valid: true,
		},
		Modified: sql.NullTime{
			Time: pbSecret.Modified.AsTime(),
			Valid: true,
		},
		Deleted: sql.NullBool{
			Bool: pbSecret.Deleted,
			Valid: true,
		},
	}
}
