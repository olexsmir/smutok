package store

// actions:
// - 'read', 'starred'
// - 'unread', 'unstar'

// SELECT article_id FROM pending_actions WHERE action = 'mark_read' ORDER BY created_at;
// DELETE FROM pending_actions WHERE action = 'mark_read' AND article_id IN ('', '?', '?');
