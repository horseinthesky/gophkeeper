package converter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gophkeeper/db/db"
	"gophkeeper/pb"
	"gophkeeper/random"
)

var tests = []struct {
	owner   string
	kind    int32
	name    string
	value   string
	deleted bool
}{
	{
		owner:   random.RandomOwner(),
		kind:    random.RandomSecretKind(),
		name:    random.RandomString(10),
		value:   random.RandomString(100),
		deleted: false,
	},
	{
		owner:   random.RandomOwner(),
		kind:    random.RandomSecretKind(),
		name:    random.RandomString(10),
		value:   random.RandomString(100),
		deleted: false,
	},
	{
		owner:   random.RandomOwner(),
		kind:    random.RandomSecretKind(),
		name:    random.RandomString(10),
		value:   random.RandomString(100),
		deleted: true,
	},
	{
		owner:   random.RandomOwner(),
		kind:    random.RandomSecretKind(),
		name:    random.RandomString(10),
		value:   random.RandomString(100),
		deleted: true,
	},
}

func TestDBSecretToPBSecret(t *testing.T) {
	for _, tt := range tests {
		now := time.Now()

		t.Run(fmt.Sprintf("test %s", tt.owner), func(t *testing.T) {
			testDBSecret := db.Secret{
				Owner: tt.owner,
				Kind:  tt.kind,
				Name:  tt.name,
				Value: []byte(tt.value),
				Created: now,
				Modified: now,
				Deleted: false,
			}

			pbSecret := DBSecretToPBSecret(testDBSecret)
			require.Equal(t, pbSecret.Owner, testDBSecret.Owner)
			require.Equal(t, pbSecret.Kind, testDBSecret.Kind)
			require.Equal(t, pbSecret.Name, testDBSecret.Name)
			require.Equal(t, pbSecret.Value, testDBSecret.Value)
			require.Equal(t, pbSecret.Created.AsTime(), testDBSecret.Created.UTC())
		})
	}
}

func TestPBSecretToBBSecret(t *testing.T) {
	for _, tt := range tests {
		now := time.Now()

		t.Run(tt.name, func(t *testing.T) {
			testPBSecret := &pb.Secret{
				Owner:    tt.owner,
				Kind:     tt.kind,
				Name:     tt.name,
				Value:    []byte(tt.value),
				Created:  timestamppb.New(now),
				Modified: timestamppb.New(now),
				Deleted:  tt.deleted,
			}

			dbSecret := PBSecretToDBSecret(testPBSecret)
			require.Equal(t, dbSecret.Owner, testPBSecret.Owner)
			require.Equal(t, dbSecret.Kind, testPBSecret.Kind)
			require.Equal(t, dbSecret.Name, testPBSecret.Name)
			require.Equal(t, dbSecret.Value, testPBSecret.Value)
			require.Equal(t, dbSecret.Created, testPBSecret.Created.AsTime())
		})
	}
}
