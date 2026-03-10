local alerts = import 'alerts/alerts.libsonnet';
local dashboards = import 'dashboards/dashboards.libsonnet';

{
  prometheusAlerts: alerts.prometheusAlerts,
  grafanaDashboards: dashboards.grafanaDashboards,
}
