package store

type StorageState string

/*
*
* Valid system states
***/
const (
	SINGLE_PRIMARY   StorageState = "single_primary"
	DUAL_WRITE       StorageState = "dual_write"
	SINGLE_SECONDARY StorageState = "single_secondary"
)
