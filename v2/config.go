package main

const (
	msgEnter       = "Entering"
	msgExit        = "Exiting"
	invitationKind = "Invitation"
	gCommitedKind  = "Committed"
	gameKind       = "Game"
	headerKind     = "Header"
	rootKind       = "Root"
	ustatsKind     = "UStats"
	noPID          = 0

	forward  direction = 1
	backward direction = -1

	// routes/paths
	idParam         = "id"
	statusParam     = "status"
	uidParam        = "uid"
	showPath        = "/show/:" + idParam
	subscribePath   = "/subscribe/:" + idParam
	unsubscribePath = "/unsubscribe/:" + idParam
	newPath         = "/new"
	undoPath        = "/undo/:" + idParam
	redoPath        = "/redo/:" + idParam
	resetPath       = "/reset/:" + idParam
	ptfinishPath    = "/ptfinish/:" + idParam
	mtfinishPath    = "/mtfinish/:" + idParam
	pfinishPath     = "/pfinish/:" + idParam
	dropPath        = "/drop/:" + idParam
	acceptPath      = "/accept/:" + idParam
	detailsPath     = "/details/:" + idParam
	updatePath      = showPath
	placeThiefPath  = "place-thief/:" + idParam
	selectThiefPath = "select-thief/:" + idParam
	moveThiefPath   = "move-thief/:" + idParam
	passPath        = "pass/:" + idParam
	playCardPath    = "play-card/:" + idParam
	msgPath         = "/message/:" + idParam
	gamePath        = "/game"
	gamesPath       = gamePath + "s"
	invitationPath  = "invitation"
	invitationsPath = invitationPath + "s"
	cuPath          = "/user/current"
	indexPath       = ":" + statusParam + "/user/:" + uidParam
	jsonIndexPath   = ":" + statusParam
	gamesIndexPath  = ":" + statusParam
	adminPath       = "/admin"
	adminGetPath    = "/:" + idParam
	adminPutPath    = adminGetPath

	gameKey        = "Game"
	jsonKey        = "JSON"
	statusKey      = "Status"
	homePath       = "/home"
	recruitingPath = "/games/recruiting"

	logKind       = "Log"
	batch   int64 = 10
)
