import { ref, computed } from 'vue';
import en from './en';
import es from './es';

// Idiomas disponibles
const languages = {
  en,
  es
};

// Idioma por defecto
const defaultLanguage = 'en';

// Idioma actual (reactivo)
const currentLanguage = ref(localStorage.getItem('language') || defaultLanguage);

// Función para cambiar el idioma
const setLanguage = (lang) => {
  if (languages[lang]) {
    currentLanguage.value = lang;
    localStorage.setItem('language', lang);
  }
};

// Traducciones actuales basadas en el idioma seleccionado
const t = computed(() => {
  return (key) => {
    return languages[currentLanguage.value][key] || languages[defaultLanguage][key] || key;
  };
});

export { currentLanguage, setLanguage, t };