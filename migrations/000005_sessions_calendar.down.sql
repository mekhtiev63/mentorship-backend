ALTER TABLE one_on_one_requests DROP CONSTRAINT IF EXISTS one_on_one_requests_calendar_event_fk;

DROP TABLE IF EXISTS final_assessments;
DROP TABLE IF EXISTS interviews;
DROP TABLE IF EXISTS calendar_event_attendees;
DROP TABLE IF EXISTS calendar_events;
DROP TABLE IF EXISTS one_on_one_requests;
