package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDToPgType(id uuid.UUID) pgtype.UUID {
	byteArray := make([]byte, 16)
	parsed, err := id.MarshalBinary()
	copy(byteArray[:], parsed[:])

	if err != nil {
		return pgtype.UUID{
			Valid: false,
		}
	}

	return pgtype.UUID{
		Bytes: [16]byte(byteArray),
		Valid: true,
	}
}

func PgTypeUUIDToUUID(pgUUID pgtype.UUID) (uuid.UUID, error) {
	if !pgUUID.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid pgtype.UUID")
	}

	return uuid.FromBytes(pgUUID.Bytes[:])
}
