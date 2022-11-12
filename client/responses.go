package client

import (
	"time"

	"github.com/benmatselby/prolificli/model"
)

// Me is a struct that represents your account.
type Me struct {
	ID                      string      `json:"id"`
	Email                   string      `json:"email"`
	DateJoined              time.Time   `json:"date_joined"`
	FirstName               string      `json:"first_name"`
	LastName                string      `json:"last_name"`
	Name                    string      `json:"name"`
	Username                string      `json:"username"`
	UserType                string      `json:"user_type"`
	CurrencyCode            string      `json:"currency_code"`
	Balance                 int         `json:"balance"`
	AvailableBalance        int         `json:"available_balance"`
	IsEmailVerified         bool        `json:"is_email_verified"`
	BillingAddress          interface{} `json:"billing_address"`
	HasPassword             bool        `json:"has_password"`
	DatetimeCreated         string      `json:"datetime_created"`
	LastLogin               string      `json:"last_login"`
	Address                 interface{} `json:"address"`
	FeesPercentage          float64     `json:"fees_percentage"`
	ServiceMarginPercentage float64     `json:"service_margin_percentage"`
	FeesPerSubmission       float64     `json:"fees_per_submission"`
	VatPercentage           float64     `json:"vat_percentage"`
	Country                 string      `json:"country"`
	ReferralURL             string      `json:"referral_url"`
	VatNumber               interface{} `json:"vat_number"`
	EmailPreferences        struct {
		BonusPayments bool `json:"bonus_payments"`
		JustPublished bool `json:"just_published"`
		Referrals     bool `json:"referrals"`
		Marketing     bool `json:"marketing"`
		PrePublish    bool `json:"pre_publish"`
	} `json:"email_preferences"`
	TermsAndConditions          bool        `json:"terms_and_conditions"`
	BetaTester                  bool        `json:"beta_tester"`
	PrivacyPolicy               bool        `json:"privacy_policy"`
	ExperimentalGroup           int         `json:"experimental_group"`
	RepresentativeSampleCredits int         `json:"representative_sample_credits"`
	RedeemableReferralCoupon    interface{} `json:"redeemable_referral_coupon"`
	MinimumRewardPerHour        int         `json:"minimum_reward_per_hour"`
	Status                      string      `json:"status"`
	OnHold                      bool        `json:"on_hold"`
	CanTopup3D                  bool        `json:"can_topup_3d"`
	HasAnsweredVatNumber        bool        `json:"has_answered_vat_number"`
	BalanceBreakdown            struct {
		ProjectFunds int `json:"project_funds"`
		ServiceFees  int `json:"service_fees"`
	} `json:"balance_breakdown"`
	CanOidcLogin                bool `json:"can_oidc_login"`
	TopupsOverReferralThreshold bool `json:"topups_over_referral_threshold"`
	IsStaff                     bool `json:"is_staff"`
	ReferralIncentive           struct {
		MinimumTopup             float64 `json:"minimum_topup"`
		RecipientCredit          float64 `json:"recipient_credit"`
		RecipientRepSampleCredit float64 `json:"recipient_rep_sample_credit"`
		ReferrerCredit           float64 `json:"referrer_credit"`
		ReferrerRepSampleCredit  int     `json:"referrer_rep_sample_credit"`
	} `json:"referral_incentive"`
	CanRunPilotStudy         bool        `json:"can_run_pilot_study"`
	NeedsToConfirmUSState    bool        `json:"needs_to_confirm_US_state"`
	MuaBetaUser              bool        `json:"mua_beta_user"`
	CurrentProjectID         interface{} `json:"current_project_id"`
	CanCashoutEnabled        bool        `json:"can_cashout_enabled"`
	CanContactSupportEnabled bool        `json:"can_contact_support_enabled"`
	CanInstantCashoutEnabled bool        `json:"can_instant_cashout_enabled"`
	InvoiceUsageEnabled      bool        `json:"invoice_usage_enabled"`
	CanOidcLoginEnabled      bool        `json:"can_oidc_login_enabled"`
	CanRunPilotStudyEnabled  bool        `json:"can_run_pilot_study_enabled"`
	Links                    struct {
		Self struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"self"`
	} `json:"_links"`
}

// ListStudiesResponse is the response for the /studies API response.
type ListStudiesResponse struct {
	Results []model.Study `json:"results"`
	Links   struct {
		Self struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"self"`
		Next struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"next"`
		Previous struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"previous"`
		Last struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"last"`
	} `json:"_links"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

// ListSubmissionsResponse is the response for the submissions request.
type ListSubmissionsResponse struct {
	Results []model.Submission `json:"results"`
	Links   struct {
		Self struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"self"`
		Next struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"next"`
		Previous struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"previous"`
		Last struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"last"`
	} `json:"_links"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

// ListRequirementsResponse is the response for the requirements request.
type ListRequirementsResponse struct {
	Results []model.Requirement `json:"results"`
	Links   struct {
		Self struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"self"`
		Next struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"next"`
		Previous struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"previous"`
		Last struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"last"`
	} `json:"_links"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

// TransitionStudyResponse is the response for transitioning a study to another status.
type TransitionStudyResponse struct {
	ID                      string        `json:"id"`
	Name                    string        `json:"name"`
	InternalName            string        `json:"internal_name"`
	Description             string        `json:"description"`
	ExternalStudyURL        string        `json:"external_study_url"`
	ProlificIDOption        string        `json:"prolific_id_option"`
	CompletionCode          string        `json:"completion_code"`
	CompletionOption        string        `json:"completion_option"`
	TotalAvailablePlaces    int           `json:"total_available_places"`
	EstimatedCompletionTime int           `json:"estimated_completion_time"`
	MaximumAllowedTime      int           `json:"maximum_allowed_time"`
	Reward                  int           `json:"reward"`
	DeviceCompatibility     []string      `json:"device_compatibility"`
	PeripheralRequirements  []interface{} `json:"peripheral_requirements"`
	EligibilityRequirements []interface{} `json:"eligibility_requirements"`
	Status                  string        `json:"status"`
}

// ListHooksResponse is the response for the hook subscriptions.
type ListHooksResponse struct {
	Results []model.Hook `json:"results"`
}

// ListHookEventTypesResponse is the response for the event types hook API.
type ListHookEventTypesResponse struct {
	Results []string `json:"results"`
}

// ListWorkspacesResponse is the response for the list workspaces endpoint.
type ListWorkspacesResponse struct {
	Results []model.Workspace `json:"results"`
}

// ListProjectsResponse is the response for the list projects endpoint.
type ListProjectsResponse struct {
	Results []model.Project `json:"results"`
}
