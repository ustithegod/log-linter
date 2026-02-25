package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	slogPath = "log/slog"
	zapPath  = "go.uber.org/zap"
)

func runWithConfig(pass *analysis.Pass, cfg Config) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		expr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		msg := getMsgExpr(pass, expr)
		if msg == nil {
			return
		}

		if cfg.Rules.LowercaseStart {
			checkForUppercase(pass, msg)
		}
		if cfg.Rules.EnglishOnly {
			checkForEnglishOnly(pass, msg)
		}
		if cfg.Rules.NoSpecials {
			checkForSpecials(pass, msg)
		}
		if cfg.Rules.Sensitive {
			checkForSensitive(pass, msg, cfg)
		}
	})

	return nil, nil
}

// getMsgExpr возвращает лог-сообщение или nil, если это не лог-вызов
func getMsgExpr(pass *analysis.Pass, call *ast.CallExpr) ast.Expr {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	// проверка по названию метода, чтобы сразу отсеять очевидные варианты
	name := sel.Sel.Name
	idx := findMsgIdx(name)
	if idx < 0 {
		return nil
	}

	var fn *types.Func

	// method (logger.Info)
	if s := pass.TypesInfo.Selections[sel]; s != nil {
		f, ok := s.Obj().(*types.Func)
		if !ok {
			return nil
		}
		fn = f
	} else {
		// package function (slog.Info)
		obj := pass.TypesInfo.Uses[sel.Sel]
		f, ok := obj.(*types.Func)
		if !ok {
			return nil
		}
		fn = f
	}

	// проверяем что это slog или zap
	if fn.Pkg() == nil {
		return nil
	}
	pkg := fn.Pkg().Path()
	if pkg != slogPath && pkg != zapPath {
		return nil
	}

	// в zap и slog метод Log имеет разную сигнатуру.
	if name == "Log" && pkg == zapPath {
		idx = 1
	}

	// проверяем что аргумент существует
	if idx >= len(call.Args) {
		return nil
	}

	return call.Args[idx]
}

// возвращает -1, если имя функции не подходит, иначе возвращает индекс аргумента msg
func findMsgIdx(name string) int {
	switch name {
	case "Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal":
		return 0
	case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
		return 1
	case "Log", "LogAttrs":
		return 2
	default:
		return -1
	}
}
