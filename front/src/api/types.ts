export type Role = "candidate" | "employer" | "admin";

export type User = {
  id: string;
  email: string;
  role: Role;
  status: "active" | "blocked" | "deleted";
  emailVerifiedAt?: string | null;
  createdAt: string;
};

export type TokenPair = {
  accessToken: string;
  refreshToken: string;
  accessExpiresAt: string;
  refreshExpiresAt: string;
  user: User;
};

export type ApiErrorBody = {
  code?: string;
  message?: string;
};

export type Named = {
  id: string;
  code?: string;
  name: string;
  rank?: number;
  industryId?: string;
  isActive?: boolean;
};

export type Category = {
  id: string;
  specializationId: string;
  gradeId: string;
  slug: string;
  isActive: boolean;
};

export type Technology = {
  id: string;
  code: string;
  name: string;
  category?: string | null;
};

export type Experience = {
  id: string;
  company: string;
  position: string;
  startedAt: string;
  endedAt?: string | null;
  description?: string | null;
};

export type CandidateTechnology = {
  technologyId: string;
  level?: number;
};

export type Profile = {
  userId: string;
  firstName?: string | null;
  lastName?: string | null;
  middleName?: string | null;
  headline?: string | null;
  about?: string | null;
  location?: string | null;
  yearsExperience?: number | null;
  categoryId?: string | null;
  gradeId?: string | null;
  specializationId?: string | null;
  fspMemberId?: string | null;
  salaryMin?: number | null;
  salaryMax?: number | null;
  salaryCurrency?: string | null;
  softSkills?: string[];
  updatedAt: string;
};

export type Contacts = {
  email?: string | null;
  phone?: string | null;
  telegram?: string | null;
  github?: string | null;
  linkedin?: string | null;
  website?: string | null;
  masked?: boolean;
  maskedFields?: string[];
};

export type FspAchievement = {
  id: string;
  eventName: string;
  eventDate?: string | null;
  place?: number | null;
  category?: string | null;
  score?: number | null;
  weight?: number;
};

export type TaskAnswer = {
  taskId: string;
  taskTitle: string;
  answer: string;
  kind: "solution" | "approach";
  userId: string;
  email: string;
  submittedAt: string;
  reaction?: "confirmed" | "thanks" | "invited" | "";
};

export type PeriodicTask = {
  id: string;
  categoryId: string;
  title: string;
  body: string;
  createdAt?: string;
  status?: "open" | "sent";
  answer?: string;
  kind?: "solution" | "approach";
  submittedAt?: string;
  reaction?: "confirmed" | "thanks" | "invited" | "";
};

export type FspState = {
  fspMemberId?: string | null;
  linkedAt?: string | null;
  achievements?: FspAchievement[];
};

export type Visibility = {
  contacts?: boolean;
  links?: boolean;
  fsp?: boolean;
  experience?: boolean;
  resume?: boolean;
  salary?: boolean;
  softSkills?: boolean;
};

export type CategoryState = {
  categoryId?: string | null;
  gradeId?: string | null;
  specializationId?: string | null;
  history?: {
    categoryId?: string;
    gradeId?: string;
    specializationId?: string;
    reason?: string;
    effectiveFrom?: string;
    effectiveTo?: string | null;
  }[];
};

export type QuestionOption = {
  id: string;
  label: string;
};

export type SessionItem = {
  id: string;
  position: number;
  type: "single_choice" | "multi_choice" | "code" | "text" | "sql" | "regex";
  topic: string;
  body: string;
  options?: QuestionOption[];
  status: "pending" | "answered" | "skipped";
};

export type TestSession = {
  id: string;
  targetCategoryId: string;
  status: string;
  startedAt: string;
  expiresAt?: string | null;
  finishedAt?: string | null;
  items?: SessionItem[];
};

export type SessionResult = {
  sessionId: string;
  score: number;
  abilityEstimate?: number;
  resultingGradeId: string;
  resultingCategoryId: string;
  decision: "confirmed" | "downgrade_offered" | "upgrade_offered";
};

export type GradeChanges = {
  canChangeAt?: string | null;
  changes?: { fromGradeId?: string | null; toGradeId: string; reason?: string; changedAt: string }[];
};

export type Invitation = {
  id: string;
  employerUserId: string;
  companyId: string;
  candidateUserId: string;
  message: string;
  salaryMin: number;
  salaryMax: number;
  currency: string;
  contactChannel: string;
  companyName?: string | null;
  employerContact?: string | null;
  status: "sent" | "viewed" | "accepted" | "rejected" | "withdrawn" | "expired";
  createdAt: string;
  statusUpdatedAt?: string;
};

export type Company = {
  id: string;
  name: string;
  description?: string | null;
  industry?: string | null;
  website?: string | null;
  size?: string | null;
};

export type Need = {
  id: string;
  title: string;
  description: string;
  categoryId?: string | null;
  specializationId?: string | null;
  gradeId?: string | null;
  stack?: string[];
  salaryMin: number;
  salaryMax: number;
  currency: string;
  workFormat?: string | null;
  status: string;
};

export type Vacancy = {
  id?: string;
  title: string;
  description: string;
  categoryId?: string;
  stack?: string[];
  salaryMin: number;
  salaryMax: number;
  workFormat?: string;
  companyId?: string;
  status?: string;
  publishedAt?: string | null;
  createdAt?: string;
};

export type SearchHit = {
  userId: string;
  displayName?: string;
  categoryId: string;
  gradeId: string;
  specializationId: string;
  testScore?: number;
  fspAchievementsCount?: number;
  fspBestPlace?: number | null;
  yearsExperience?: number | null;
  score: number;
  explanation?: string | null;
  reasons: string[];
};

export type SearchPage = {
  items: SearchHit[];
  total: number;
  limit: number;
  offset: number;
};

export type CandidateCard = {
  userId: string;
  displayName: string;
  headline?: string | null;
  categoryId?: string | null;
  gradeId?: string | null;
  specializationId?: string | null;
  yearsExperience?: number | null;
  location?: string | null;
  fsp?: { hasFSP: boolean; achievementsCount: number; bestPlace?: number | null };
  contacts?: Contacts;
};

export type Application = {
  id: string;
  candidateUserId: string;
  vacancyId: string;
  companyId: string;
  coverLetter?: string | null;
  status: string;
  createdAt: string;
};

export type Resume = {
  id: string;
  title: string;
  source: string;
  parsed?: Record<string, unknown> | null;
  createdAt: string;
};
