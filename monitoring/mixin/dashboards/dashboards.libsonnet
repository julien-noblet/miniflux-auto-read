local config = import '../config.libsonnet';

{
  grafanaDashboards+:: {
    'miniflux-auto-read.json': {
      uid: config._config.dashboards.uid,
      title: config._config.dashboards.title,
      tags: ['miniflux'],
      timezone: 'browser',
      schemaVersion: 26,
      panels: [
        {
          title: 'Entries Processed',
          type: 'graph',
          datasource: 'Prometheus',
          targets: [
            { expr: 'sum(rate(miniflux_entries_processed_total{%(minifluxAutoReadSelector)s}[5m]))' % config._config, legendFormat: 'Processed' },
          ],
          gridPos: { h: 8, w: 12, x: 0, y: 0 },
        },
        {
          title: 'Processing Errors',
          type: 'graph',
          datasource: 'Prometheus',
          targets: [
            { expr: 'sum(rate(miniflux_entries_processing_errors_total{%(minifluxAutoReadSelector)s}[5m])) by (type)' % config._config, legendFormat: '{{type}}' },
          ],
          gridPos: { h: 8, w: 12, x: 12, y: 0 },
        },
        {
          title: 'API Latency (95th percentile)',
          type: 'graph',
          datasource: 'Prometheus',
          targets: [
            { expr: 'histogram_quantile(0.95, sum(rate(miniflux_api_duration_seconds_bucket{%(minifluxAutoReadSelector)s}[5m])) by (le, operation))' % config._config, legendFormat: '{{operation}}' },
          ],
          gridPos: { h: 8, w: 24, x: 0, y: 8 },
        },
      ],
    },
  },
}
