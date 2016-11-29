package sync

type SyncSessionStreamWrapper struct {
	Client SyncService_SyncSessionClient
	Server SyncService_SyncSessionServer
}
