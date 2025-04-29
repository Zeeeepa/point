# Point: UI Mockup Specifications

This document provides detailed specifications for the UI mockups of the Point project. These mockups serve as a visual guide for the implementation of the user interface.

## Dashboard

### Dashboard Overview

![Dashboard Overview](https://via.placeholder.com/800x500?text=Dashboard+Overview)

#### Components:
1. **Header**
   - Logo
   - Navigation menu
   - User profile
   - Settings button
   - Notifications icon

2. **Status Panel**
   - Active providers count
   - Active endpoints count
   - Request rate (requests/minute)
   - Success rate
   - Average response time

3. **Quick Actions**
   - Add Provider button
   - Test System button
   - View Logs button
   - Import/Export button

4. **Recent Activity Feed**
   - Timestamp
   - Activity type (request, configuration change, error)
   - Brief description
   - Status indicator

5. **System Health**
   - Component status indicators
   - Resource usage (CPU, memory)
   - Uptime

## Provider Management

### Provider List

![Provider List](https://via.placeholder.com/800x500?text=Provider+List)

#### Components:
1. **Filter Bar**
   - Search box
   - Provider type filter (API, Web Interface)
   - Status filter (Active, Inactive)
   - Sort options

2. **View Toggle**
   - List view
   - Card view

3. **Provider Cards**
   - Provider name
   - Provider type icon
   - Status indicator
   - Quick actions (Edit, Delete, Test, Enable/Disable)
   - Models count
   - Last tested timestamp

4. **Add Provider Button**
   - Floating action button

### API Provider Configuration

![API Provider Configuration](https://via.placeholder.com/800x500?text=API+Provider+Configuration)

#### Components:
1. **Provider Selection**
   - Provider type dropdown
   - Provider logo display

2. **Basic Information**
   - Custom name input
   - API key input (masked)
   - Base URL input
   - Organization ID input

3. **Model Configuration**
   - Auto-fetch models button
   - Model list with checkboxes
   - Model mapping interface
   - Add custom model button

4. **Advanced Settings**
   - Collapsible panel
   - Headers configuration
   - Proxy settings
   - Rate limiting
   - Timeout settings

5. **Action Buttons**
   - Save
   - Test
   - Cancel

### Web Interface Configuration

![Web Interface Configuration](https://via.placeholder.com/800x500?text=Web+Interface+Configuration)

#### Components:
1. **Interface Selection**
   - Interface type dropdown
   - Interface logo display

2. **Basic Information**
   - Custom name input
   - URL input (for custom interfaces)

3. **Authentication**
   - Authentication method selection
   - Credentials inputs
   - Session persistence toggle

4. **Browser Settings**
   - Browser type selection
   - Headless mode toggle
   - User data directory input
   - Timeout settings

5. **Advanced Settings**
   - Collapsible panel
   - Custom CSS selectors
   - JavaScript injection
   - Screenshot settings

6. **Action Buttons**
   - Save
   - Test
   - Cancel

### Provider Testing

![Provider Testing](https://via.placeholder.com/800x500?text=Provider+Testing)

#### Components:
1. **Provider Information**
   - Provider name
   - Provider type
   - Status

2. **Test Input**
   - Model selection
   - Prompt input
   - Parameters (temperature, max tokens, etc.)

3. **Test Results**
   - Response display
   - Response time
   - Token usage
   - Cost estimate

4. **Logs Panel**
   - Request details
   - Response details
   - Error messages (if any)

5. **Action Buttons**
   - Run Test
   - Save Results
   - Close

## Endpoint Configuration

### Endpoint Selection

![Endpoint Selection](https://via.placeholder.com/800x500?text=Endpoint+Selection)

#### Components:
1. **Endpoint List**
   - Endpoint name
   - Toggle switch
   - Status indicator
   - Request count
   - Configure button

2. **Endpoint Details**
   - Endpoint description
   - Compatible providers
   - Required parameters
   - Example request

3. **Action Buttons**
   - Save Changes
   - Reset

### Provider Mapping

![Provider Mapping](https://via.placeholder.com/800x500?text=Provider+Mapping)

#### Components:
1. **Endpoint Selection**
   - Endpoint dropdown or tabs

2. **Primary Provider**
   - Provider selection dropdown
   - Model pattern input
   - Test button

3. **Fallback Providers**
   - Ordered list of providers
   - Add/remove buttons
   - Reorder buttons

4. **Routing Options**
   - Load balancing method selection
   - Advanced routing toggle

5. **Model-Specific Routing**
   - Model pattern input
   - Provider selection
   - Add/remove buttons

6. **Action Buttons**
   - Save
   - Test
   - Cancel

### Routing Rules Builder

![Routing Rules Builder](https://via.placeholder.com/800x500?text=Routing+Rules+Builder)

#### Components:
1. **Rule List**
   - Rule name
   - Rule conditions summary
   - Rule actions summary
   - Enable/disable toggle
   - Edit/delete buttons

2. **Rule Editor**
   - Rule name input
   - Condition builder
     - Condition type dropdown
     - Operator dropdown
     - Value input
     - Add/remove condition buttons
   - Action builder
     - Action type dropdown
     - Provider selection
     - Parameters
     - Add/remove action buttons

3. **Rule Testing**
   - Test input
   - Test results
   - Matched rules display

4. **Action Buttons**
   - Save Rule
   - Test Rule
   - Cancel

## Configuration Management

### Configuration Profiles

![Configuration Profiles](https://via.placeholder.com/800x500?text=Configuration+Profiles)

#### Components:
1. **Profile List**
   - Profile name
   - Description
   - Created date
   - Last modified date
   - Default indicator
   - Action buttons (Load, Edit, Delete, Set as Default)

2. **Profile Details**
   - Profile name
   - Description
   - Provider count
   - Endpoint count
   - Rule count

3. **Action Buttons**
   - Save Current as New
   - Import
   - Export
   - Compare

### System Settings

![System Settings](https://via.placeholder.com/800x500?text=System+Settings)

#### Components:
1. **Settings Tabs**
   - Server
   - Authentication
   - Logging
   - Monitoring

2. **Server Settings**
   - Port input
   - Host input
   - SSL toggle and certificate inputs
   - CORS settings

3. **Authentication Settings**
   - Authentication toggle
   - User management
   - API key management

4. **Logging Settings**
   - Log level selection
   - Log rotation settings
   - Log storage location

5. **Monitoring Settings**
   - Metrics collection toggle
   - Alert thresholds
   - Notification channels

6. **Action Buttons**
   - Save
   - Reset
   - Test

## Monitoring and Logs

### Request Monitor

![Request Monitor](https://via.placeholder.com/800x500?text=Request+Monitor)

#### Components:
1. **Filter Bar**
   - Time range selection
   - Provider filter
   - Endpoint filter
   - Status filter
   - Search box

2. **Request List**
   - Timestamp
   - Endpoint
   - Provider
   - Model
   - Status code
   - Response time
   - Expand button

3. **Request Details**
   - Request headers
   - Request body
   - Response headers
   - Response body
   - Performance metrics

4. **Action Buttons**
   - Refresh
   - Export
   - Clear

### Usage Statistics

![Usage Statistics](https://via.placeholder.com/800x500?text=Usage+Statistics)

#### Components:
1. **Time Range Selection**
   - Preset ranges (hour, day, week, month)
   - Custom range picker

2. **Metrics Selection**
   - Requests
   - Tokens
   - Cost
   - Response time

3. **Provider Usage Chart**
   - Bar or pie chart
   - Provider breakdown
   - Legend

4. **Endpoint Usage Chart**
   - Bar or pie chart
   - Endpoint breakdown
   - Legend

5. **Time Series Chart**
   - Line chart
   - Time-based usage
   - Multiple series

6. **Export Options**
   - CSV
   - Excel
   - JSON

### System Logs

![System Logs](https://via.placeholder.com/800x500?text=System+Logs)

#### Components:
1. **Filter Bar**
   - Log level selection
   - Component filter
   - Time range selection
   - Search box

2. **Log List**
   - Timestamp
   - Log level indicator
   - Component
   - Message
   - Expand button

3. **Log Details**
   - Full message
   - Stack trace (if error)
   - Context data
   - Related logs

4. **Action Buttons**
   - Refresh
   - Export
   - Clear

## Mobile Responsive Design

### Mobile Dashboard

![Mobile Dashboard](https://via.placeholder.com/400x800?text=Mobile+Dashboard)

#### Components:
1. **Header**
   - Logo
   - Menu button
   - Notifications icon

2. **Status Summary**
   - Condensed status indicators
   - Collapsible panels

3. **Quick Actions**
   - Icon buttons
   - Action menu

4. **Recent Activity**
   - Scrollable list
   - Simplified entries

### Mobile Provider Management

![Mobile Provider Management](https://via.placeholder.com/400x800?text=Mobile+Provider+Management)

#### Components:
1. **Provider List**
   - Stacked cards
   - Simplified information
   - Swipe actions

2. **Provider Configuration**
   - Stepped form
   - Progressive disclosure
   - Simplified inputs

## Design System

### Color Palette

- **Primary**: #3f51b5
- **Secondary**: #f50057
- **Success**: #4caf50
- **Warning**: #ff9800
- **Error**: #f44336
- **Info**: #2196f3
- **Background**: #f5f5f5
- **Surface**: #ffffff
- **Text Primary**: #212121
- **Text Secondary**: #757575

### Typography

- **Font Family**: Roboto, sans-serif
- **Headings**:
  - H1: 24px, 400 weight
  - H2: 20px, 500 weight
  - H3: 18px, 500 weight
  - H4: 16px, 500 weight
- **Body**:
  - Body 1: 16px, 400 weight
  - Body 2: 14px, 400 weight
- **Caption**: 12px, 400 weight

### Components

- **Buttons**:
  - Primary: Filled, rounded corners
  - Secondary: Outlined, rounded corners
  - Text: No background, no border
  - Icon: Circular, with tooltip

- **Cards**:
  - Elevation: Light shadow
  - Border Radius: 4px
  - Padding: 16px

- **Forms**:
  - Inputs: Outlined style
  - Labels: Floating
  - Validation: Inline messages
  - Spacing: 16px between fields

- **Tables**:
  - Header: Sticky, bold text
  - Rows: Alternating background
  - Pagination: Bottom right
  - Row Actions: Icon buttons

## Implementation Notes

1. **Accessibility**:
   - All components should meet WCAG 2.1 AA standards
   - Proper contrast ratios
   - Keyboard navigation
   - Screen reader support

2. **Responsiveness**:
   - Desktop-first design
   - Breakpoints:
     - Small: 0-600px
     - Medium: 600-960px
     - Large: 960-1280px
     - Extra Large: 1280px+

3. **Performance**:
   - Lazy loading for complex components
   - Virtualized lists for large datasets
   - Optimized images
   - Code splitting

4. **Animation**:
   - Subtle transitions between states
   - Loading indicators
   - Feedback animations

5. **Dark Mode**:
   - Full dark mode support
   - System preference detection
   - Manual toggle

## Next Steps

1. Create high-fidelity mockups based on these specifications
2. Develop component prototypes
3. Conduct usability testing
4. Refine designs based on feedback
5. Implement UI components according to the design system

