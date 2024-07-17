package osrscache

const (
	TierEnumID         = 3967
	TypeEnumID         = 3968
	TierMapEnumID      = 3980
	MonsterEnumID      = 3970
	CAIDParamID        = 1306
	TitleParamID       = 1308
	DescriptionParamID = 1309
	TierParamID        = 1310
	TypeParamID        = 1311
	MonsterParamID     = 1312
)

type CombatAchievementTier string

const (
	TierEasy        CombatAchievementTier = "Easy"
	TierMedium      CombatAchievementTier = "Medium"
	TierHard        CombatAchievementTier = "Hard"
	TierElite       CombatAchievementTier = "Elite"
	TierMaster      CombatAchievementTier = "Master"
	TierGrandmaster CombatAchievementTier = "Grandmaster"
)

type CombatAchievementType string

const (
	TypeStamina     CombatAchievementType = "Stamina"
	TypePerfection  CombatAchievementType = "Perfection"
	TypeKillCount   CombatAchievementType = "Kill Count"
	TypeMechanical  CombatAchievementType = "Mechanical"
	TypeRestriction CombatAchievementType = "Restriction"
	TypeSpeed       CombatAchievementType = "Speed"
)

type CombatAchievement struct {
	ID          uint32                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Monster     string                `json:"monster"`
	Tier        CombatAchievementTier `json:"tier"`
	Type        CombatAchievementType `json:"type"`
}
