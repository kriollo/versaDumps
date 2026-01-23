import { describe, it, expect, beforeEach } from 'vitest';
import { currentLanguage, setLanguage, t } from '../index';

describe('i18n', () => {
  beforeEach(() => {
    // Reset to default language before each test
    setLanguage('en');
  });

  it('has default language set to en', () => {
    expect(currentLanguage.value).toBe('en');
  });

  it('changes language to Spanish', () => {
    setLanguage('es');
    expect(currentLanguage.value).toBe('es');
  });

  it('changes language to English', () => {
    setLanguage('en');
    expect(currentLanguage.value).toBe('en');
  });

  it('translates common keys in English', () => {
    setLanguage('en');
    
    expect(t.value('settings')).toBeTruthy();
    expect(t.value('server')).toBeTruthy();
    expect(t.value('port')).toBeTruthy();
    expect(t.value('language')).toBeTruthy();
    expect(t.value('save')).toBeTruthy();
    expect(t.value('close')).toBeTruthy();
  });

  it('translates common keys in Spanish', () => {
    setLanguage('es');
    
    expect(t.value('settings')).toBeTruthy();
    expect(t.value('server')).toBeTruthy();
    expect(t.value('port')).toBeTruthy();
    expect(t.value('language')).toBeTruthy();
    expect(t.value('save')).toBeTruthy();
    expect(t.value('close')).toBeTruthy();
  });

  it('provides different translations for different languages', () => {
    setLanguage('en');
    const englishSettings = t.value('settings');
    
    setLanguage('es');
    const spanishSettings = t.value('settings');
    
    expect(englishSettings).not.toBe(spanishSettings);
  });

  it('persists language to localStorage', () => {
    const mockLocalStorage = {
      data: {},
      setItem(key, value) {
        this.data[key] = value;
      },
      getItem(key) {
        return this.data[key] || null;
      },
    };
    
    global.localStorage = mockLocalStorage;
    
    setLanguage('es');
    expect(mockLocalStorage.getItem('language')).toBe('es');
    
    setLanguage('en');
    expect(mockLocalStorage.getItem('language')).toBe('en');
  });

  it('has profile-related translations', () => {
    setLanguage('en');
    
    expect(t.value('profiles')).toBeTruthy();
    expect(t.value('active_profile')).toBeTruthy();
    expect(t.value('create_profile')).toBeTruthy();
    expect(t.value('delete_profile')).toBeTruthy();
  });

  it('has log-related translations', () => {
    setLanguage('en');
    
    expect(t.value('log_files')).toBeTruthy();
    expect(t.value('log_folders')).toBeTruthy();
    expect(t.value('clear_logs')).toBeTruthy();
  });

  it('has error level translations', () => {
    setLanguage('en');
    
    expect(t.value('error')).toBeTruthy();
    expect(t.value('warning')).toBeTruthy();
    expect(t.value('info')).toBeTruthy();
    expect(t.value('debug')).toBeTruthy();
    expect(t.value('success')).toBeTruthy();
  });

  it('handles missing translation keys gracefully', () => {
    setLanguage('en');
    
    const result = t.value('non_existent_key_12345');
    // Should return the key itself or some default value
    expect(result).toBeTruthy();
  });

  it('maintains reactivity when language changes', () => {
    setLanguage('en');
    const translation1 = t.value('settings');
    
    setLanguage('es');
    const translation2 = t.value('settings');
    
    // Both should be truthy but different
    expect(translation1).toBeTruthy();
    expect(translation2).toBeTruthy();
    expect(translation1).not.toBe(translation2);
  });
});
