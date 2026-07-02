package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httputil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
)

var (
	parseStrFuncMap = map[reflect.Type]func(string) (any, error){
		reflect.TypeFor[time.Time](): func(s string) (any, error) {
			return time.Parse(time.RFC3339, s)
		},
		reflect.TypeFor[time.Duration](): func(s string) (any, error) {
			return time.ParseDuration(s)
		},
		reflect.TypeFor[timeutil.Date](): func(s string) (any, error) {
			return timeutil.ParseDate(s)
		},
		reflect.TypeFor[timeutil.Duration](): func(s string) (any, error) {
			return timeutil.ParseDuration(s)
		},
		reflect.TypeFor[unit.DataSize](): func(s string) (any, error) {
			return unit.ParseDataSizeString(s)
		},
	}

	defaultParseFuncs = []mapstructure.DecodeHookFunc{
		// TODO: add more parse hook functions for specific types

		func(f reflect.Type, t reflect.Type, data any) (any, error) {
			switch f.Kind() { //nolint:exhaustive
			case reflect.String:
				if fun, exists := parseStrFuncMap[t]; exists {
					return fun(data.(string)) //nolint:forcetypeassert
				}
				return data, nil
			default:
				return data, nil
			}
		},

		// NOTE: Parse slice should be the last hook function as some
		// types may be treated as slice, such as UUID can be []byte.
		mapstructure.StringToSliceHookFunc(","),
	}
)

var (
	wsUpgrader = &websocket.Upgrader{
		Subprotocols: []string{"access_token"},
		CheckOrigin:  func(r *http.Request) bool { return true },
	}
)

// BaseHandler base handler for every other handler to inherit from
type BaseHandler struct {
	sysErrorUC *syserroruc.UC
}

func New(
	sysErrorUC *syserroruc.UC,
) *BaseHandler {
	return &BaseHandler{
		sysErrorUC: sysErrorUC,
	}
}

func (h *BaseHandler) RequestCtx(ginCtx *gin.Context) context.Context {
	// TODO: for now just use the input context without any conversion
	return ginCtx
}

// RenderResponse renders response to client as JSON
func (h *BaseHandler) RenderResponse(ctx *gin.Context, status int, body any) {
	ctx.JSON(status, body)
}

// RenderError renders errors to client as JSON
func (h *BaseHandler) RenderError(ctx *gin.Context, err error) {
	// Parse the error
	errInfo, _ := apperrors.ParseError(err, h.ParseRequestLang(ctx))
	h.SaveError(ctx, err, errInfo)

	// Remove the error debug data from the response if we are not in dev env
	if !config.Current.IsDevEnv() {
		// these fields are for dev only
		errInfo.Cause = ""
		errInfo.DebugLog = ""
		errInfo.StackTrace = ""

		// Also remove the details from InnerErrors
		for i := range errInfo.InnerErrors {
			errInfo.InnerErrors[i].Cause = ""
		}
	}

	ctx.JSON(errInfo.Status, errInfo)
}

// SaveError save error in to DB
func (h *BaseHandler) SaveError(ctx *gin.Context, _ error, errInfo *apperrors.ErrorInfo) {
	_, _ = h.sysErrorUC.CreateSysError(ctx, &syserrordto.CreateSysErrorReq{
		ErrorInfo: errInfo,
	})
}

// parsePagination parses paging and sorting params
//
//nolint:gocognit
func (h *BaseHandler) parsePagination(ctx *gin.Context, paging *basedto.Paging) error {
	if paging.Limit == 0 {
		paging.Limit = basedto.PageLimitDefault
	}

	if limitStr := ctx.Query("pageLimit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > basedto.PageLimitMax {
			return apperrors.NewArgumentInvalid("pageLimit")
		}
		paging.Limit = limit
	}
	if offsetStr := ctx.Query("pageOffset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return apperrors.NewArgumentInvalid("pageOffset")
		}
		if offset > 0 {
			paging.Offset = offset
		}
	}

	//nolint:nestif
	if sortQuery := ctx.Query("sort"); sortQuery != "" {
		orders := basedto.Orders{}
		sort := strings.Split(sortQuery, ",")
		for _, str := range sort {
			if strings.HasPrefix(str, "-") {
				if len(str) == 1 {
					return apperrors.NewArgumentInvalid("sort")
				}
				orders.Add(&basedto.Order{Direction: basedto.DirectionDesc, ColumnName: str[1:]})
			} else {
				orders.Add(&basedto.Order{Direction: basedto.DirectionAsc, ColumnName: str})
			}
		}
		if len(paging.Sort) > 0 {
			orders.Add(paging.Sort...)
		}
		paging.Sort = orders
	}
	// Converts camelCase to snake_case
	for _, order := range paging.Sort {
		// Column name may include table such as `task.created_at`
		parts := strings.Split(order.ColumnName, ".")
		for i := range parts {
			parts[i] = strutil.ToSnakeCase(parts[i])
		}
		order.ColumnName = strings.Join(parts, ".")
	}

	return nil
}

// parseQuery parse request query from URL of style /api?key1=val1&key2=val2
func (h *BaseHandler) parseQuery(ctx *gin.Context, query any) error {
	mapQuery := map[string]string{}
	urlQueries := ctx.Request.URL.Query()
	for k, v := range urlQueries {
		mapQuery[k] = strings.Join(v, ",")
	}
	// NOTE: caller should call ParseMultipartForm()/ParseForm() if the request contains form data.
	for formParam, value := range ctx.Request.Form {
		mapQuery[formParam] = strings.Join(value, ",")
	}
	if len(mapQuery) == 0 {
		return nil
	}

	config := &mapstructure.DecoderConfig{
		Result:           query,
		ZeroFields:       true,
		WeaklyTypedInput: true,
		Squash:           true,
	}

	if len(defaultParseFuncs) == 1 {
		config.DecodeHook = defaultParseFuncs[0]
	} else if len(defaultParseFuncs) > 1 {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(defaultParseFuncs...)
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return apperrors.New(err)
	}
	if err = decoder.Decode(mapQuery); err != nil {
		return apperrors.NewArgumentInvalid("query").WithCause(err)
	}

	return nil
}

// ParseAndValidateRequest parses request query, then validate the request.
//
// See ParseRequest.
func (h *BaseHandler) ParseAndValidateRequest(ctx *gin.Context, reqStruct any, paging *basedto.Paging) error {
	err := h.ParseRequest(ctx, reqStruct, paging)
	if err != nil {
		return apperrors.New(err)
	}

	// Execute custom modifier for the request input
	if modifier, ok := reqStruct.(basedto.ReqModifier); ok {
		if err = modifier.ModifyRequest(); err != nil {
			return apperrors.New(err)
		}
	}

	// Execute custom validator for the request input
	if validator, ok := reqStruct.(basedto.ReqValidator); ok {
		if vldErrs := validator.Validate(); len(vldErrs) > 0 {
			return vldErrs
		}
	}

	return nil
}

// ParseRequest parses request query.
//
// The request struct should be defined similarly as:
//
//	type ListProjectReq struct {
//	   Key1 int    `mapstructure:"key1"`
//	   Key2 string `mapstructure:"key2"`
//	}
func (h *BaseHandler) ParseRequest(ctx *gin.Context, reqStruct any, paging *basedto.Paging) error {
	if err := h.parseQuery(ctx, reqStruct); err != nil {
		return err
	}
	if paging != nil {
		if err := h.parsePagination(ctx, paging); err != nil {
			return err
		}
	}
	return nil
}

// ParseStringParam parse string from context param
func (h *BaseHandler) ParseStringParam(ctx *gin.Context, paramName string) (string, error) {
	idStr := ctx.Params.ByName(paramName)
	if idStr == "" {
		return "", apperrors.New(apperrors.ErrBadRequest).
			WithMsgLog("require param `%s` of type string in URL", paramName)
	}
	return idStr, nil
}

// ParseIntParam parse uint from context param
func (h *BaseHandler) ParseIntParam(ctx *gin.Context, paramName string) (int, error) {
	value, err := strconv.ParseInt(ctx.Param(paramName), 10, 64) //nolint:mnd
	if err != nil {
		return 0, apperrors.New(apperrors.ErrBadRequest).WithCause(err).
			WithMsgLog("require param `%s` of type int in URL", paramName)
	}
	return int(value), nil
}

// ParseUintParam parse uint from context param
func (h *BaseHandler) ParseUintParam(ctx *gin.Context, paramName string) (uint, error) {
	value, err := strconv.ParseUint(ctx.Param(paramName), 10, 64) //nolint:mnd
	if err != nil {
		return 0, apperrors.New(apperrors.ErrBadRequest).WithCause(err).
			WithMsgLog("require param `%s` of type uint in URL", paramName)
	}
	return uint(value), nil
}

// ParseAndValidateJSONBody parse request body as JSON, then validate the result
func (h *BaseHandler) ParseAndValidateJSONBody(ctx *gin.Context, reqStruct any) error {
	err := h.ParseJSONBody(ctx, reqStruct)
	if err != nil {
		return apperrors.New(err)
	}

	// Execute custom modifier for the request input
	if modifier, ok := reqStruct.(basedto.ReqModifier); ok {
		if err = modifier.ModifyRequest(); err != nil {
			return apperrors.New(err)
		}
	}

	// Execute custom validator for the request input
	if validator, ok := reqStruct.(basedto.ReqValidator); ok {
		if vldErrs := validator.Validate(); len(vldErrs) > 0 {
			return vldErrs
		}
	}

	return nil
}

// ParseJSONBody parse request body as JSON
func (h *BaseHandler) ParseJSONBody(ctx *gin.Context, reqStruct any) error {
	var buf bytes.Buffer
	ctx.Request.Body = io.NopCloser(io.TeeReader(ctx.Request.Body, &buf))

	if err := ctx.ShouldBindJSON(reqStruct); err != nil && buf.Len() > 0 {
		if handler, ok := reqStruct.(basedto.ReqParsingErrorHandler); ok {
			if newErr := handler.HandleParsingError(err); newErr != nil {
				return apperrors.New(newErr)
			}
		}
		return apperrors.New(apperrors.ErrBadRequest).WithCause(err)
	}

	ctx.Request.Body = io.NopCloser(&buf)
	return nil
}

func (h *BaseHandler) ParseRequestLang(ctx *gin.Context) translation.Lang {
	return httputil.ParseRequestLang(ctx.GetHeader("Accept-Language"))
}

func (h *BaseHandler) StreamAppLogs(
	ctx *gin.Context,
	staticLogs []*tasklog.LogFrame, // static logs are in DB
	realtimeLogsStream <-chan []*tasklog.LogFrame, // realtime logs streaming channel
	logsStreamCloser func() error,
) {
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer conn.Close()

	if logsStreamCloser != nil {
		defer func() {
			_ = logsStreamCloser()
		}()
	}

	writeFrames := func(frames []*tasklog.LogFrame) error {
		dataBytes, err := json.Marshal(frames)
		if err != nil {
			return apperrors.New(err)
		}
		return conn.WriteMessage(websocket.BinaryMessage, dataBytes)
	}

	// Send static logs first
	for _, chunk := range gofn.Chunk(staticLogs, 100) { //nolint:mnd
		if err := writeFrames(chunk); err != nil {
			return
		}
	}

	// Read loop to detect client connection close/abort
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Stream real-time logs
	if realtimeLogsStream != nil {
		for {
			select {
			case <-done:
				return
			case log, ok := <-realtimeLogsStream:
				if !ok {
					return
				}
				if err := writeFrames(log); err != nil {
					return
				}
			}
		}
	}
}

func (h *BaseHandler) UpgradeWebsocket(ctx *gin.Context) (*websocket.Conn, error) {
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return conn, nil
}

func (h *BaseHandler) IsWebsocketRequest(ctx *gin.Context) bool {
	return strings.ToLower(ctx.Request.Header.Get("Connection")) == "upgrade"
}

func (h *BaseHandler) ParseFormFiles(ctx *gin.Context, req *filedto.UploadReq) error {
	cfg := &config.Current.Files
	err := ctx.Request.ParseMultipartForm(int64(cfg.RequestMaxSize))
	if err != nil {
		if errors.Is(err, multipart.ErrMessageTooLarge) {
			return apperrors.New(apperrors.ErrRequestTooBig).WithParam("MaxSize", cfg.RequestMaxSize)
		}
		return apperrors.New(err)
	}

	fileType := ctx.PostForm("fileType")
	var maxFile int
	var maxFileSize unit.DataSize
	var fileExts []string
	var requiredScopes []base.ObjectScopeType
	switch base.FileType(fileType) {
	case base.FileTypeBuildSource:
		maxFile = cfg.BuildSourceMaxFile
		maxFileSize = cfg.BuildSourceMaxSize
		fileExts = cfg.BuildSourceFileExts
		requiredScopes = []base.ObjectScopeType{base.ObjectScopeApp}
	case base.FileTypeSystemBackup, base.FileTypeRepoCache:
		fallthrough
	case base.FileTypeSchedJobOutput:
		fallthrough
	default:
		return apperrors.New(apperrors.ErrFileTypeNotSupported).
			WithParam("SupportedTypes", []base.FileType{base.FileTypeBuildSource})
	}
	req.FileType = base.FileType(fileType)

	scope := base.ObjectScopeType(ctx.PostForm("scope"))
	if !gofn.Contain(requiredScopes, scope) {
		return apperrors.New(apperrors.ErrFileScopeUnsupported).WithParam("Scope", scope)
	}

	switch scope {
	case base.ObjectScopeApp:
		projectID, appID := ctx.PostForm("projectId"), ctx.PostForm("appId")
		if projectID == "" || appID == "" {
			return apperrors.New(apperrors.ErrParamMissing).WithParam("Name", "projectId or appId")
		}
		req.Scope = base.NewObjectScopeApp(appID, projectID)
	case base.ObjectScopeProject:
		projectID := ctx.PostForm("projectId")
		if projectID == "" {
			return apperrors.New(apperrors.ErrParamMissing).WithParam("Name", "projectId")
		}
		req.Scope = base.NewObjectScopeProject(projectID)
	case base.ObjectScopeUser:
		userID := ctx.PostForm("userId")
		if userID == "" {
			return apperrors.New(apperrors.ErrParamMissing).WithParam("Name", "userId")
		}
		req.Scope = base.NewObjectScopeUser(userID)
	case base.ObjectScopeGlobal, "global":
		req.Scope = base.NewObjectScopeGlobal()
	default:
		return apperrors.New(apperrors.ErrObjectScopeInvalid).WithParam("Scope", scope)
	}

	req.StorageType = base.FileStorageType(ctx.PostForm("storageType"))
	if !gofn.Contain(base.AllFileStorageTypes, req.StorageType) {
		return apperrors.New(apperrors.ErrStorageTypeUnsupported).WithParam("Type", req.StorageType)
	}
	req.StorageID = ctx.PostForm("storageId")

	form, err := ctx.MultipartForm()
	if err != nil {
		return apperrors.New(err)
	}
	if maxFile > 0 && len(form.File["file"]) > maxFile {
		return apperrors.New(apperrors.ErrTooMany).WithParam("Name", "Files").
			WithNTParam("MaxItem", maxFile)
	}
	allowAnyExt := gofn.Contain(fileExts, "*")
	for _, formFile := range form.File["file"] {
		if maxFileSize > 0 && formFile.Size > maxFileSize.Bytes() {
			return apperrors.New(apperrors.ErrFileSizeTooBig).
				WithNTParam("MaxSize", maxFileSize)
		}
		if !(allowAnyExt || gofn.Contain(fileExts, strings.ToLower(filepath.Ext(formFile.Filename)))) { //nolint
			return apperrors.New(apperrors.ErrFileExtNotSupported).WithNTParam("SupportedExts", fileExts)
		}
		if cfg.FileNameMaxLength > 0 && gofn.RuneLength(formFile.Filename) > cfg.FileNameMaxLength {
			return apperrors.New(apperrors.ErrFileNameTooLong).WithNTParam("MaxNameLen", cfg.FileNameMaxLength)
		}
		req.Files = append(req.Files, formFile)
	}

	return nil
}
