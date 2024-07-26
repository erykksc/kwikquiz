package pastgames

import (
	"github.com/erykksc/kwikquiz/internal/common"
)

var pastGameTmpl = common.ParseTmplWithFuncs("templates/pastgames/pastgame.html")
var pastGamesListTmpl = common.TmplParseWithBase("templates/pastgames/search_pastgames.html")
