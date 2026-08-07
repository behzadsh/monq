package example

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//go:generate go run github.com/behzadsh/monq/cmd/monqgen -type sessionDoc

// sessionDoc is the second document of this example and is unexported, the way a type describing what a collection
// stores usually is when it never leaves its package.
//
// It is here because that case names the generated value after the model as written, sessionDocPaths, which is the
// name the root type would take as well. The file next to this one is what the generator writes when it has to keep
// the two apart, and the build compiling it is the check that it managed.
type sessionDoc struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    bson.ObjectID `bson:"user_id"`
	Device    device        `bson:"device"`
	ExpiresAt time.Time     `bson:"expires_at"`
}

// device is a nested document of [sessionDoc].
type device struct {
	Platform string `bson:"platform"`
	Version  string `bson:"version"`
}
