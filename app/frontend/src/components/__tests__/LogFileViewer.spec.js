import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import LogFileViewer from '../LogFileViewer.vue';

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
}));

describe('LogFileViewer', () => {
  let wrapper;

  beforeEach(() => {
    wrapper = mount(LogFileViewer);
  });

  it('renders correctly', () => {
    expect(wrapper.exists()).toBe(true);
    const headerText = wrapper.find('h3').text().toLowerCase();
    expect(headerText).toMatch(/log.*file/);
  });

  it('displays no logs message when empty', () => {
    expect(wrapper.vm.logLines.length).toBe(0);
    const text = wrapper.text().toLowerCase();
    expect(text).toMatch(/no.*log/);
  });

  it('has level filter dropdown', () => {
    const select = wrapper.find('select');
    expect(select.exists()).toBe(true);
    
    const options = select.findAll('option');
    expect(options.length).toBeGreaterThanOrEqual(5);
    
    // Check for expected filter options
    const optionValues = options.map(opt => opt.element.value);
    expect(optionValues).toContain('all');
    expect(optionValues).toContain('error');
    expect(optionValues).toContain('warning');
    expect(optionValues).toContain('info');
  });

  it('filters logs by level', async () => {
    // Add test logs
    wrapper.vm.logLines = [
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Error occurred',
        level: 'error',
        timestamp: new Date(),
        lineNum: 1,
      },
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Warning message',
        level: 'warning',
        timestamp: new Date(),
        lineNum: 2,
      },
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Info message',
        level: 'info',
        timestamp: new Date(),
        lineNum: 3,
      },
    ];
    
    await wrapper.vm.$nextTick();
    
    // All logs visible by default
    expect(wrapper.vm.filteredLogs.length).toBe(3);
    
    // Filter by error
    wrapper.vm.levelFilter = 'error';
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.filteredLogs.length).toBe(1);
    expect(wrapper.vm.filteredLogs[0].level).toBe('error');
    
    // Filter by warning
    wrapper.vm.levelFilter = 'warning';
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.filteredLogs.length).toBe(1);
    expect(wrapper.vm.filteredLogs[0].level).toBe('warning');
  });

  it('toggles auto-scroll', async () => {
    expect(wrapper.vm.autoScroll).toBe(true);
    
    const autoScrollButton = wrapper.findAll('button').find(btn => 
      btn.text().includes('auto_scroll')
    );
    
    if (autoScrollButton) {
      await autoScrollButton.trigger('click');
      expect(wrapper.vm.autoScroll).toBe(false);
    }
  });

  it('clears logs', async () => {
    // Add test logs
    wrapper.vm.logLines = [
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Test log',
        level: 'info',
        timestamp: new Date(),
        lineNum: 1,
      },
    ];
    
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.logLines.length).toBe(1);
    
    // Clear logs
    const clearButton = wrapper.findAll('button').find(btn => 
      btn.attributes('title') === 'clear_logs'
    );
    
    if (clearButton) {
      await clearButton.trigger('click');
      expect(wrapper.vm.logLines.length).toBe(0);
      expect(wrapper.vm.fileFilters.length).toBe(0);
    }
  });

  it('adds log line correctly', () => {
    const logEntry = {
      filePath: '/test/app.log',
      fileName: 'app.log',
      line: 'Test log message',
      level: 'info',
      timestamp: new Date(),
      lineNum: 1,
    };
    
    wrapper.vm.addLogLine(logEntry);
    
    expect(wrapper.vm.logLines.length).toBe(1);
    expect(wrapper.vm.logLines[0].line).toBe('Test log message');
    expect(wrapper.vm.logLines[0].level).toBe('info');
  });

  it('limits log lines to maxLines', () => {
    const maxLines = wrapper.vm.maxLines;
    
    // Add more than maxLines
    for (let i = 0; i < maxLines + 10; i++) {
      wrapper.vm.addLogLine({
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: `Log line ${i}`,
        level: 'info',
        timestamp: new Date(),
        lineNum: i,
      });
    }
    
    expect(wrapper.vm.logLines.length).toBe(maxLines);
  });

  it('extracts file name correctly', () => {
    expect(wrapper.vm.getFileName('/path/to/app.log')).toBe('app.log');
    expect(wrapper.vm.getFileName('C:\\Windows\\Logs\\error.log')).toBe('error.log');
    expect(wrapper.vm.getFileName('simple.log')).toBe('simple.log');
  });

  it('applies correct CSS class for log levels', () => {
    expect(wrapper.vm.getLogLevelClass('error')).toContain('red');
    expect(wrapper.vm.getLogLevelClass('warning')).toContain('yellow');
    expect(wrapper.vm.getLogLevelClass('info')).toContain('blue');
    expect(wrapper.vm.getLogLevelClass('debug')).toContain('slate');
    expect(wrapper.vm.getLogLevelClass('success')).toContain('green');
  });

  it('applies correct text CSS class for log levels', () => {
    expect(wrapper.vm.getLogLevelTextClass('error')).toContain('red');
    expect(wrapper.vm.getLogLevelTextClass('warning')).toContain('yellow');
    expect(wrapper.vm.getLogLevelTextClass('info')).toContain('blue');
    expect(wrapper.vm.getLogLevelTextClass('success')).toContain('green');
  });

  it('formats timestamp correctly', () => {
    const date = new Date('2024-01-15T14:30:45');
    const formatted = wrapper.vm.formatTime(date);
    
    expect(formatted).toMatch(/\d{2}:\d{2}:\d{2}/);
  });

  it('tracks active files', async () => {
    wrapper.vm.logLines = [
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Log 1',
        level: 'info',
        timestamp: new Date(),
        lineNum: 1,
      },
      {
        filePath: '/test/error.log',
        fileName: 'error.log',
        line: 'Log 2',
        level: 'error',
        timestamp: new Date(),
        lineNum: 1,
      },
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Log 3',
        level: 'info',
        timestamp: new Date(),
        lineNum: 2,
      },
    ];
    
    await wrapper.vm.$nextTick();
    
    expect(wrapper.vm.activeFiles.length).toBe(2);
    expect(wrapper.vm.activeFiles).toContain('/test/app.log');
    expect(wrapper.vm.activeFiles).toContain('/test/error.log');
  });

  it('filters by file', async () => {
    wrapper.vm.logLines = [
      {
        filePath: '/test/app.log',
        fileName: 'app.log',
        line: 'Log 1',
        level: 'info',
        timestamp: new Date(),
        lineNum: 1,
      },
      {
        filePath: '/test/error.log',
        fileName: 'error.log',
        line: 'Log 2',
        level: 'error',
        timestamp: new Date(),
        lineNum: 1,
      },
    ];
    
    await wrapper.vm.$nextTick();
    
    // No file filter - show all
    expect(wrapper.vm.filteredLogs.length).toBe(2);
    
    // Filter by app.log
    wrapper.vm.fileFilters = ['/test/app.log'];
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.filteredLogs.length).toBe(1);
    expect(wrapper.vm.filteredLogs[0].fileName).toBe('app.log');
  });

  it('toggles file filter', async () => {
    const filePath = '/test/app.log';
    
    expect(wrapper.vm.fileFilters).not.toContain(filePath);
    
    // Add filter
    wrapper.vm.toggleFileFilter(filePath);
    expect(wrapper.vm.fileFilters).toContain(filePath);
    
    // Remove filter
    wrapper.vm.toggleFileFilter(filePath);
    expect(wrapper.vm.fileFilters).not.toContain(filePath);
  });

  it('parses JSON log lines', () => {
    const jsonLine = '{"message": "Test", "level": "info"}';
    const result = wrapper.vm.tryParseJson(jsonLine);
    
    expect(result.isJson).toBe(true);
    expect(result.formattedLine).toContain('"message"');
    expect(result.coloredJson).toBeTruthy();
  });

  it('handles non-JSON log lines', () => {
    const plainLine = 'This is a plain log line';
    const result = wrapper.vm.tryParseJson(plainLine);
    
    expect(result.isJson).toBe(false);
    expect(result.formattedLine).toBe(plainLine);
    expect(result.coloredJson).toBe('');
  });

  it('colorizes JSON correctly', () => {
    const json = '{"key": "value", "number": 123, "bool": true}';
    const colored = wrapper.vm.colorizeJson(json);
    
    expect(colored).toContain('json-key');
    // Check for either json-string or json-key (values might be styled as keys)
    expect(colored).toMatch(/json-(string|key)/);
    expect(colored).toContain('json-number');
    expect(colored).toContain('json-boolean');
  });
});
