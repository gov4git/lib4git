package id

import (
	"context"

	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/git"
	"github.com/gov4git/lib4git/must"
)

func Init(
	ctx context.Context,
	ownerAddr OwnerAddress,
) git.Change[form.None, PrivateCredentials] {
	ownerCloned := CloneOwner(ctx, ownerAddr)
	privChg := InitLocal(ctx, ownerAddr, ownerCloned)

	ownerCloned.Public.Push(ctx)
	ownerCloned.Private.Push(ctx)
	return privChg
}

func InitLocal(
	ctx context.Context,
	ownerAddr OwnerAddress,
	ownerCloned OwnerCloned,
) git.Change[form.None, PrivateCredentials] {

	privChg := initPrivateStageOnly(ctx, ownerCloned.Private.Tree(), ownerAddr)
	pubChg := initPublicStageOnly(ctx, ownerCloned.Public.Tree(), privChg.Result.PublicCredentials)
	git.Commit(ctx, ownerCloned.Private.Tree(), privChg.Msg)
	git.Commit(ctx, ownerCloned.Public.Tree(), pubChg.Msg)
	return privChg
}

func initPrivateStageOnly(ctx context.Context, priv *git.Tree, ownerAddr OwnerAddress) git.Change[form.None, PrivateCredentials] {
	if _, err := priv.Filesystem.Stat(PrivateCredentialsNS.OSPath()); err == nil {
		must.Errorf(ctx, "private credentials file already exists")
	}
	cred, err := GenerateCredentials()
	must.NoError(ctx, err)
	git.ToFileStage(ctx, priv, PrivateCredentialsNS.OSPath(), cred)
	return git.NewChange(
		"Initialized private credentials.",
		"id_init_private",
		form.None{},
		cred,
		nil,
	)
}

func initPublicStageOnly(ctx context.Context, pub *git.Tree, cred PublicCredentials) git.ChangeNoResult {
	if _, err := pub.Filesystem.Stat(PublicCredentialsNS.OSPath()); err == nil {
		must.Errorf(ctx, "public credentials file already exists")
	}
	git.ToFileStage(ctx, pub, PublicCredentialsNS.OSPath(), cred)
	return git.NewChangeNoResult("Initialized public credentials.", "id_init_public")
}
