// Package supabase implements the core persistence ports on top of a
// Postgres/Supabase database, using pgx directly with hand-written SQL.
package supabase

import (
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

// registerUUIDType teaches a pgtype.Map how to encode and scan the standard
// library uuid.UUID type, which pgx does not support natively. It mirrors the
// wrapper-plan approach pgx's third-party UUID integrations use.
func registerUUIDType(m *pgtype.Map) {
	m.TryWrapEncodePlanFuncs = append(
		[]pgtype.TryWrapEncodePlanFunc{tryWrapUUIDEncodePlan},
		m.TryWrapEncodePlanFuncs...,
	)
	m.TryWrapScanPlanFuncs = append(
		[]pgtype.TryWrapScanPlanFunc{tryWrapUUIDScanPlan},
		m.TryWrapScanPlanFuncs...,
	)
}

// wrapUUID adapts uuid.UUID to pgtype's UUIDValuer / UUIDScanner interfaces.
type wrapUUID struct {
	uuid uuid.UUID
}

func (w wrapUUID) UUIDValue() (pgtype.UUID, error) {
	return pgtype.UUID{Bytes: [16]byte(w.uuid), Valid: true}, nil
}

func (w *wrapUUID) ScanUUID(v pgtype.UUID) error {
	if !v.Valid {
		w.uuid = uuid.UUID{}
		return nil
	}
	w.uuid = uuid.UUID(v.Bytes)
	return nil
}

func tryWrapUUIDEncodePlan(value any) (pgtype.WrappedEncodePlanNextSetter, any, bool) {
	if v, ok := value.(uuid.UUID); ok {
		return &wrapUUIDEncodePlan{}, wrapUUID{uuid: v}, true
	}
	return nil, nil, false
}

type wrapUUIDEncodePlan struct {
	next pgtype.EncodePlan
}

func (p *wrapUUIDEncodePlan) SetNext(next pgtype.EncodePlan) { p.next = next }

func (p *wrapUUIDEncodePlan) Encode(value any, buf []byte) ([]byte, error) {
	return p.next.Encode(wrapUUID{uuid: value.(uuid.UUID)}, buf)
}

func tryWrapUUIDScanPlan(target any) (pgtype.WrappedScanPlanNextSetter, any, bool) {
	if _, ok := target.(*uuid.UUID); ok {
		return &wrapUUIDScanPlan{}, &wrapUUID{}, true
	}
	return nil, nil, false
}

type wrapUUIDScanPlan struct {
	next pgtype.ScanPlan
}

func (p *wrapUUIDScanPlan) SetNext(next pgtype.ScanPlan) { p.next = next }

func (p *wrapUUIDScanPlan) Scan(src []byte, dst any) error {
	w := &wrapUUID{}
	if err := p.next.Scan(src, w); err != nil {
		return err
	}
	*(dst.(*uuid.UUID)) = w.uuid
	return nil
}
