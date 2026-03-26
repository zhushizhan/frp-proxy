# Requirements

## Summary

新增一个并行于现有 `frpc` Dashboard 的 `webui` 前端入口，复用现有 Store API 和 Config API，为代理与访客提供更便捷的类型化创建和编辑体验，同时保留原有经典界面与通用配置编辑功能。

## Requirements

### Requirement 1: Parallel webui entry

**User Story:** As a frpc administrator, I want a new `/webui/` interface alongside the existing dashboard, so that I can adopt the new guided workflow without losing access to the classic UI.

#### Acceptance Criteria

1. When `frpc` starts with the web assets embedded, the existing classic dashboard remains accessible at its current route.
2. When `frpc` starts with the web assets embedded, the new guided UI is accessible under `/webui/`.
3. When a user refreshes or deep-links to a `webui` route, the SPA loads correctly instead of returning a 404.
4. The new `webui` static assets do not conflict with the classic dashboard assets.

### Requirement 2: Reuse existing Store API only

**User Story:** As a frpc administrator, I want the new UI to save structured entries through the existing Store API, so that the new interface changes only the operator experience, not the persistence model.

#### Acceptance Criteria

1. Creating, editing, deleting, and toggling Store-managed proxies in `webui` uses the existing `/api/store/proxies` endpoints.
2. Creating, editing, and deleting Store-managed visitors in `webui` uses the existing `/api/store/visitors` endpoints.
3. The new UI does not add any new configuration writeback API, managed-config API, or companion-file persistence.
4. Store-managed entries created from `webui` remain visible and editable from the classic UI.

### Requirement 3: Preserve classic config editing

**User Story:** As a frpc administrator, I want the new UI to keep the existing raw config editor, so that I can still inspect and edit global client configuration directly when needed.

#### Acceptance Criteria

1. The `webui` contains a Config page that reads from the existing `/api/config` endpoint.
2. Saving from the `webui` Config page writes via the existing `/api/config` endpoint and reloads through `/api/reload`.
3. The structured proxy and visitor flows do not directly rewrite the main config file outside of existing Store API behavior.

### Requirement 4: Guided proxy creation and editing

**User Story:** As a frpc administrator, I want a guided proxy workflow organized by proxy type, so that I can create common proxy types faster and with less configuration guesswork.

#### Acceptance Criteria

1. The `webui` exposes guided entry cards for `tcp`, `udp`, `http`, `https`, `tcpmux`, `stcp`, `sudp`, and `xtcp`.
2. Each type entry shows a short description, typical use case, and an important caution where applicable.
3. Creating a proxy uses a two-step experience: choose type first, then fill in the corresponding form.
4. Editing a proxy opens directly into the form for its existing type.
5. The proxy form prioritizes high-frequency required fields and hides irrelevant sections for the selected type.
6. Advanced options remain available without removing current Store API field coverage.

### Requirement 5: Guided visitor creation and editing

**User Story:** As a frpc administrator, I want a guided visitor workflow organized by visitor type, so that I can create paired private-access flows more confidently.

#### Acceptance Criteria

1. The `webui` exposes guided entry cards for `stcp`, `sudp`, and `xtcp` visitors.
2. Visitor creation uses a two-step experience: choose type first, then fill in the corresponding form.
3. The UI explicitly explains the relationship between private-access proxy types and visitor types.
4. Editing a visitor opens directly into the form for its existing type.

### Requirement 6: Store-disabled behavior

**User Story:** As a frpc administrator, I want clear guidance when Store API is not enabled, so that I understand why structured creation is unavailable and what to configure.

#### Acceptance Criteria

1. When Store API is disabled, `webui` shows a clear explanation and an example of the required `[store]` configuration.
2. When Store API is disabled, guided create/edit actions are not allowed.
3. When Store API is disabled, the Config page remains usable.

### Requirement 7: Interface orientation and switching

**User Story:** As a frpc administrator, I want the guided UI to feel intentionally optimized for common operations and still let me return to the classic UI, so that I can choose the interface that best fits my current task.

#### Acceptance Criteria

1. The proxy and visitor landing pages in `webui` prioritize creation by type over generic status browsing.
2. The `webui` prominently identifies itself as the guided interface.
3. The `webui` provides a link back to the classic interface.
4. The classic interface is not required to add a link into `webui` in this feature.
