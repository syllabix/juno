-- +migrate Up
ALTER TABLE [dbo].[UserSessions]
DROP FK_UserID, COLUMN [UserID];

ALTER TABLE [dbo].[UserSessions]
ADD [ContentsJSON] NVARCHAR(MAX) NULL;

