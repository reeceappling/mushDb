package aaa_temp_rbac

/*
Custom Role-Based Access Controls with SQL

Users Table
id=username

Consider making the next 3 tables all one table, but add a column that is either "role","org", or "project"(
	Role Users Table
	id, roleId, userId, permissionLevel (0 is read, 1 is writeRoleItems, 2 is editRoleButNotChangePerms?, 3 is roleAdmin?)

	Org users table
	id, roleId, userId, permissionLevel (0 is read, 1 is writeRoleItems, 2 is editRoleButNotChangePerms?, 3 is roleAdmin?)

	Project users table
	id, project, userId, permissionLevel (0 is read, 1 is writeProjectItems, 2 is editProjectButNotChangePerms, 3 is projectAdmin)
) described On next 2 lines:
Group Table
id, userId, groupType (role/org/project), role/org/project ID, permissionLevel (0 is read, 1 is writeGroupItems, 2 is editGroupButNotChangePerms?, 3 is groupAdmin?)


*/
/*
Custom Authentication stuff with SQL

Users Table
id=username

Email Table
id, userId, email # Each user can have multiple emails

Passwords table
id, userId, salt, hashedPass

Google table
id, userId, googleId? (any others?)

2FA tables (
	Phone nums table
	id, userId, phoneNum

	Phone nums table
	id, userId, phoneNum

	TOTP Table? the server stores a unique secret key (or seed) associated with the user's account, usually as a Base32 string. It also stores configuration details like the time-step size (usually 30 seconds) and the hashing algorithm (such as HMAC-SHA1)
	Passkey table?
		Biometrics
		DeviceTouchId, FaceScans, IrisScans,
	HardwareKey table?

	PushNotifications2FaTable?
)



*/
