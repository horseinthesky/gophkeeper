package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"gophkeeper/db/db"
	"gophkeeper/pb"
)

func DBSecretToPBSecret(secret db.Secret) *pb.Secret {
	return &pb.Secret{
		Owner:    secret.Owner,
		Kind:     secret.Kind,
		Name:     secret.Name,
		Value:    secret.Value,
		Created:  timestamppb.New(secret.Created),
		Modified: timestamppb.New(secret.Modified),
		Deleted:  secret.Deleted,
	}
}

func PBSecretToDBSecret(secret *pb.Secret) db.Secret {
	return db.Secret{
		Owner: secret.Owner,
		Kind:  secret.Kind,
		Name:  secret.Name,
		Value: secret.Value,
		Created: secret.Created.AsTime(),
		Modified: secret.Modified.AsTime(),
		Deleted: secret.Deleted,
	}
}
