package basesettinghandler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc/notificationdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc/oauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc/sslproviderdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

// Every list ListSetting serves reads the page asked for: a request that did
// not would be answered whole, whatever pageLimit says.
func TestEverySettingsListPages(t *testing.T) {
	for _, req := range []any{
		accesstokendto.NewListAccessTokenReq(),
		acmednsproviderdto.NewListAcmeDnsProviderReq(),
		apikeydto.NewListAPIKeyReq(),
		backuprepodto.NewListBackupRepoReq(),
		basicauthdto.NewListBasicAuthReq(),
		cloudstoragedto.NewListCloudStorageReq(),
		commandpipedto.NewListCommandPipeReq(),
		commandtemplatedto.NewListCommandTemplateReq(),
		configfiledto.NewListConfigFileReq(),
		emaildto.NewListEmailReq(),
		githubappdto.NewListGithubAppReq(),
		imservicedto.NewListIMServiceReq(),
		networkdto.NewListNetworkReq(),
		nodedto.NewListNodeReq(),
		notificationdto.NewListNotificationReq(),
		oauthdto.NewListOAuthReq(),
		periodicjobdto.NewListPeriodicJobReq(),
		registryauthdto.NewListRegistryAuthReq(),
		repowebhookdto.NewListRepoWebhookReq(),
		schedjobdto.NewListSchedJobReq(),
		secretdto.NewListSecretReq(),
		settingmountdto.NewListSettingMountReq(),
		sshkeydto.NewListSSHKeyReq(),
		sslcertdto.NewListSSLCertReq(),
		sslproviderdto.NewListSSLProviderReq(),
		volumedto.NewListVolumeReq(),
	} {
		_, ok := req.(interface{ PagingReq() *basedto.Paging })
		assert.True(t, ok, "%T does not page", req)
	}
}
