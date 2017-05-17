-- +migrate Up
-- NOTE: This is an example migration script which you should copy into your project,
-- should you wish to use this package.

CREATE TABLE [dbo].[UserRoles] (
    [RoleID] INT NOT NULL IDENTITY (1,1),
    [RoleName] NVARCHAR(50) NOT NULL,
    [Created] DATETIMEOFFSET NOT NULL
        CONSTRAINT [DF_RoleCreated] DEFAULT (SYSDATETIMEOFFSET()),
    CONSTRAINT [PK_RoleID] PRIMARY KEY ([RoleID]),
    CONSTRAINT [UQ_RoleName] UNIQUE ([RoleName])    
);

CREATE TABLE [dbo].[Permissions] (
    [PermissionID] INT NOT NULL IDENTITY (1,1),
    [Label] NVARCHAR(50) NOT NULL,
    [Description] NVARCHAR(255) NULL,
    [Created] DATETIMEOFFSET NOT NULL
        CONSTRAINT [DF_PermissionCreated] DEFAULT (SYSDATETIMEOFFSET()),
    CONSTRAINT [PK_PermissionID] PRIMARY KEY ([PermissionID]),
    CONSTRAINT [UQ_PermLabel] UNIQUE ([Label]),    
);

CREATE TABLE [dbo].[UserRolePermissionsMap] (
    [RoleID] INT NOT NULL,
    [PermissionID] INT NOT NULL,
    CONSTRAINT [PK_RolePermission] PRIMARY KEY (RoleID, PermissionID),
    CONSTRAINT [FK_MapRoleID] FOREIGN KEY ([RoleID]) REFERENCES dbo.UserRoles([RoleID]),
    CONSTRAINT [FK_MapPermissionID] FOREIGN KEY ([PermissionID]) REFERENCES dbo.Permissions([PermissionID])
);

CREATE TABLE [dbo].[Users] (
    [UserID] INT NOT NULL IDENTITY (1,1),
    [Email] nvarchar(255) NOT NULL,    
    [Password] nvarchar(60) NOT NULL,
    [RoleID] INT NOT NULL,
    [Created] DATETIMEOFFSET NOT NULL
        CONSTRAINT [DF_UserCreated] DEFAULT (SYSDATETIMEOFFSET()),
    [Modified] DATETIMEOFFSET NOT NULL,
    [LastLogin] DATETIMEOFFSET NULL,
    CONSTRAINT [PK_UserID] PRIMARY KEY ([UserID]),
    CONSTRAINT [UQ_UserEmail] UNIQUE ([Email]),
    CONSTRAINT [FK_UserRoleID] FOREIGN KEY ([RoleID]) REFERENCES dbo.UserRoles([RoleID])
);

CREATE TABLE [dbo].[UserSessions] (
    [GUID] UNIQUEIDENTIFIER NOT NULL, 
    [UserID] INT NULL,    
    [StartTime] DATETIMEOFFSET 
        NOT NULL CONSTRAINT [DF_SessionStartTime] DEFAULT (SYSDATETIMEOFFSET()),
    [Expiration] DATETIMEOFFSET NOT NULL,
    CONSTRAINT [PK_GUID] PRIMARY KEY NONCLUSTERED (GUID),
    CONSTRAINT [FK_UserID] FOREIGN KEY ([UserID]) REFERENCES dbo.Users([UserID])
);
