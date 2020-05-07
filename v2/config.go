package main

const (
	msgEnter       = "Entering"
	msgExit        = "Exiting"
	invitationKind = "Invitation"
	gameKind       = "Game"
	historyKind    = "History"
	gheaderKind    = "GHeader"
	rootKind       = "Root"
	noPID          = 0

	forward  = 1
	backward = -1

	// routes/paths
	idParam         = "id"
	statusParam     = "status"
	uidParam        = "uid"
	showPath        = "/show/:" + idParam
	newPath         = "/new"
	undoPath        = "/undo/:" + idParam
	redoPath        = "/redo/:" + idParam
	resetPath       = "/reset/:" + idParam
	ptfinishPath    = "/ptfinish/:" + idParam
	mtfinishPath    = "/mtfinish/:" + idParam
	dropPath        = "/drop/:" + idParam
	acceptPath      = "/accept/:" + idParam
	updatePath      = showPath
	placeThiefPath  = "place-thief/:" + idParam
	selectThiefPath = "select-thief/:" + idParam
	moveThiefPath   = "move-thief/:" + idParam
	playCardPath    = "play-card/:" + idParam
	msgPath         = showPath + "/addmessage"
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
	homePath       = "/"
	recruitingPath = "/games/recruiting"

	logKind       = "Log"
	batch   int64 = 10
)
