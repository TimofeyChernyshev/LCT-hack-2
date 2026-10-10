package http

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type item struct {
	api.SessionItem
	Options []option `json:"options,omitempty"`
	correct string
}

type session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CategoryID uuid.UUID
	GradeID    uuid.UUID
	Status     string
	StartedAt  time.Time
	ExpiresAt  time.Time
	Items      []item
	Answers    map[uuid.UUID]bool
}

type Server struct {
	mu       sync.Mutex
	sessions map[uuid.UUID]*session
	tasks    []miniTask
	answers  []miniAnswer
}

type miniTask struct {
	ID         uuid.UUID
	CategoryID uuid.UUID
	Title      string
	Body       string
	CreatedAt  time.Time
}

type miniAnswer struct {
	TaskID      uuid.UUID
	UserID      uuid.UUID
	Email       string
	Answer      string
	Kind        string
	SubmittedAt time.Time
	Reaction    string
}

func NewServer() *Server {
	return &Server{sessions: map[uuid.UUID]*session{}}
}

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

func (s *Server) ListPeriodicTasks(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	uid, _ := uuid.Parse(user.ID)
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]gin.H, 0, len(s.tasks))
	for _, task := range s.tasks {
		item := gin.H{
			"id": task.ID, "categoryId": task.CategoryID, "title": task.Title,
			"body": task.Body, "createdAt": task.CreatedAt, "status": "open",
		}
		for _, answer := range s.answers {
			if answer.TaskID == task.ID && answer.UserID == uid {
				item["status"] = "sent"
				item["answer"] = answer.Answer
				item["kind"] = answer.Kind
				item["submittedAt"] = answer.SubmittedAt
				item["reaction"] = answer.Reaction
			}
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) SubmitPeriodicTask(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "candidate" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "отвечает соискатель")
		return
	}
	var body struct {
		Answer string `json:"answer"`
		Kind   string `json:"kind"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Answer) == "" {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужен ответ")
		return
	}
	if body.Kind != "approach" {
		body.Kind = "solution"
	}
	uid, _ := uuid.Parse(user.ID)
	taskID := uuid.UUID(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	found := false
	for _, task := range s.tasks {
		if task.ID == taskID {
			found = true
			break
		}
	}
	if !found {
		httpx.GinError(c, http.StatusNotFound, "not_found", "задание не найдено")
		return
	}
	next := miniAnswer{TaskID: taskID, UserID: uid, Email: user.Email, Answer: strings.TrimSpace(body.Answer), Kind: body.Kind, SubmittedAt: time.Now()}
	replaced := false
	for i := range s.answers {
		if s.answers[i].TaskID == taskID && s.answers[i].UserID == uid {
			s.answers[i] = next
			replaced = true
			break
		}
	}
	if !replaced {
		s.answers = append(s.answers, next)
	}
	c.Status(http.StatusOK)
}

func (s *Server) CreatePeriodicTask(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "создаёт работодатель")
		return
	}
	var body api.PeriodicTaskInput
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Title) == "" || strings.TrimSpace(body.Body) == "" {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужны название и текст")
		return
	}
	now := time.Now()
	task := miniTask{ID: uuid.New(), CategoryID: uuid.UUID(body.CategoryId), Title: strings.TrimSpace(body.Title), Body: strings.TrimSpace(body.Body), CreatedAt: now}
	s.mu.Lock()
	s.tasks = append(s.tasks, task)
	s.mu.Unlock()
	c.JSON(http.StatusCreated, api.PeriodicTask{
		Id: &task.ID, CategoryId: body.CategoryId, Title: task.Title, Body: task.Body, CreatedAt: &now,
	})
}

func (s *Server) ListTaskResponses(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "список ответов для работодателя")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]gin.H, 0, len(s.answers))
	for _, answer := range s.answers {
		title := ""
		for _, task := range s.tasks {
			if task.ID == answer.TaskID {
				title = task.Title
				break
			}
		}
		out = append(out, gin.H{
			"taskId": answer.TaskID, "taskTitle": title, "userId": answer.UserID,
			"answer": answer.Answer, "kind": answer.Kind, "email": answer.Email,
			"submittedAt": answer.SubmittedAt, "reaction": answer.Reaction,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ReactToAnswer(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "реагирует работодатель")
		return
	}
	var body struct {
		TaskID   string `json:"taskId"`
		UserID   string `json:"userId"`
		Reaction string `json:"reaction"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужна реакция")
		return
	}
	switch body.Reaction {
	case "confirmed", "thanks", "invited":
	default:
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "неизвестная реакция")
		return
	}
	taskID, errTask := uuid.Parse(body.TaskID)
	userID, errUser := uuid.Parse(body.UserID)
	if errTask != nil || errUser != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужны задание и соискатель")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.answers {
		if s.answers[i].TaskID == taskID && s.answers[i].UserID == userID {
			s.answers[i].Reaction = body.Reaction
			c.Status(http.StatusOK)
			return
		}
	}
	httpx.GinError(c, http.StatusNotFound, "not_found", "ответ не найден")
}

func (s *Server) InternalGetCategory(c *gin.Context, userID openapi_types.UUID) {
	httpx.GinError(c, http.StatusNotFound, "not_found", "категория не найдена")
}
func (s *Server) ListMySessions(c *gin.Context) { c.JSON(http.StatusOK, []api.Session{}) }

func (s *Server) StartSession(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	var body struct {
		TargetCategoryId openapi_types.UUID `json:"targetCategoryId"`
		GradeId          openapi_types.UUID `json:"gradeId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.TargetCategoryId == uuid.Nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужна категория")
		return
	}
	uid, _ := uuid.Parse(user.ID)
	now := time.Now()
	gradeID := uuid.UUID(body.GradeId)
	sess := &session{
		ID: uuid.New(), UserID: uid, CategoryID: uuid.UUID(body.TargetCategoryId), GradeID: gradeID,
		Status: "in_progress", StartedAt: now, ExpiresAt: now.Add(20 * time.Minute), Answers: map[uuid.UUID]bool{},
		Items: questions(),
	}
	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()
	c.JSON(http.StatusCreated, sess.view(false))
}

func (s *Server) GetSession(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	sess, ok := s.own(c, id, user.ID)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, sess.view(true))
}

func (s *Server) AnswerItem(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	sess, ok := s.own(c, id, user.ID)
	if !ok {
		return
	}
	var body api.AnswerInput
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужен ответ")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range sess.Items {
		if sess.Items[i].Id == body.ItemId {
			sess.Items[i].Status = api.Answered
			sess.Answers[uuid.UUID(body.ItemId)] = answerMatches(sess.Items[i], body.Answer)
			c.Status(http.StatusOK)
			return
		}
	}
	httpx.GinError(c, http.StatusNotFound, "not_found", "вопрос не найден")
}

func (s *Server) SubmitSession(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	sess, ok := s.own(c, id, user.ID)
	if !ok {
		return
	}
	s.mu.Lock()
	sess.Status = "evaluated"
	correct := 0
	for _, ok := range sess.Answers {
		if ok {
			correct++
		}
	}
	s.mu.Unlock()
	decision := api.DowngradeOffered
	if correct >= 2 {
		decision = api.Confirmed
	}
	score := float32(correct) / 3 * 100
	c.JSON(http.StatusOK, api.SessionResult{
		SessionId: sess.ID, Score: score, ResultingCategoryId: sess.CategoryID, ResultingGradeId: sess.GradeID, Decision: decision,
	})
}

func (s *Server) ListGradeChanges(c *gin.Context) {
	if _, ok := mustUser(c); !ok {
		return
	}
	next := time.Now().Add(90 * 24 * time.Hour)
	c.JSON(http.StatusOK, gin.H{"canChangeAt": next, "changes": []api.GradeChange{}})
}

func (s *Server) own(c *gin.Context, id openapi_types.UUID, userID string) (*session, bool) {
	uid, _ := uuid.Parse(userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.sessions[uuid.UUID(id)]
	if sess == nil || sess.UserID != uid {
		httpx.GinError(c, http.StatusNotFound, "not_found", "сессия не найдена")
		return nil, false
	}
	return sess, true
}

func (s *session) view(withItems bool) gin.H {
	body := gin.H{
		"id": s.ID, "targetCategoryId": s.CategoryID, "status": s.Status,
		"startedAt": s.StartedAt, "expiresAt": s.ExpiresAt,
	}
	if withItems {
		body["items"] = s.Items
	}
	return body
}

func questions() []item {
	return []item{
		choice(1, "Индексы", "Что даёт B-tree индекс в базе данных?", "a", []option{{"a", "Ускоряет поиск по ключу"}, {"b", "Шифрует таблицу"}, {"c", "Удаляет строки"}}),
		choice(2, "HTTP", "Какой код ответа значит «создано»?", "a", []option{{"a", "201"}, {"b", "404"}, {"c", "500"}}),
		{
			SessionItem: api.SessionItem{
				Id: uuid.New(), Position: 3, Type: api.Text, Topic: "Функция",
				Body: "Что вернёт вызов add(2, 3)? Напишите только число.\n\nfunc add(a, b int) int {\n  return a + b\n}",
				Status: api.Pending,
			},
			correct: "5",
		},
	}
}

func choice(position int, topic, body, correct string, options []option) item {
	return item{
		SessionItem: api.SessionItem{Id: uuid.New(), Position: position, Type: api.SingleChoice, Topic: topic, Body: body, Status: api.Pending},
		Options:     options,
		correct:     correct,
	}
}

func answerMatches(it item, answer any) bool {
	text, ok := answer.(string)
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(text), it.correct)
}

func NewRouter(si api.ServerInterface, signer jwtx.Signer, origins []string, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{AllowOrigins: origins, AllowMethods: []string{"GET", "POST", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type"}, MaxAge: 5 * time.Minute}))
	w := &api.ServerInterfaceWrapper{Handler: si, ErrorHandler: func(c *gin.Context, err error, code int) {
		c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
	}}
	r.GET("/healthz", w.Healthz)
	auth := r.Group("")
	auth.Use(httpx.GinJWTAuth(signer), httpx.GinRequireVerifiedEmail())
	auth.POST("/me/sessions", w.StartSession)
	auth.GET("/me/sessions", w.ListMySessions)
	auth.GET("/me/sessions/:id", w.GetSession)
	auth.POST("/me/sessions/:id/answers", w.AnswerItem)
	auth.POST("/me/sessions/:id/submit", w.SubmitSession)
	auth.GET("/me/category-changes", w.ListGradeChanges)
	auth.GET("/me/periodic-tasks", w.ListPeriodicTasks)
	auth.POST("/me/periodic-tasks/:id/submit", w.SubmitPeriodicTask)
	auth.POST("/periodic-tasks", w.CreatePeriodicTask)
	if srv, ok := si.(*Server); ok {
		auth.GET("/periodic-tasks", srv.ListTaskResponses)
		auth.POST("/periodic-tasks/reactions", srv.ReactToAnswer)
	}
	r.GET("/internal/users/:userId/category", w.InternalGetCategory)
	return r
}

func mustUser(c *gin.Context) (httpx.UserInfo, bool) {
	user, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
	}
	return user, ok
}
