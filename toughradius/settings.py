#!/usr/bin/env python
#coding=utf-8
import os, sys

__version__ = '5.0'
__license__ = 'Apache License 2.0'

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

abspath = lambda *p: os.path.abspath(os.path.join(*p))
PROJECT_ROOT = abspath(os.path.dirname(__file__))
GUARDIAN_MODULE_PATH = abspath(PROJECT_ROOT, '..')
sys.path.insert(0, GUARDIAN_MODULE_PATH)


SECRET_KEY = '+indr81v&1m0t0utzx#al9ef$*8ea3s-#=&-4q=-z49@m^eelu'

DEBUG = True

ALLOWED_HOSTS = ['127.0.0.1']

LOGGING = {
    'version': 1,
    'disable_existing_loggers': False,
    'handlers': {
        'console': {
            'class': 'logging.StreamHandler',
        },
    },
    'loggers': {
        'django': {
            'handlers': ['console'],
            'level': os.getenv('DJANGO_LOG_LEVEL', 'INFO'),
        },
    },
}



# Application definition

INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',
    'guardian',
    'toughradius.apps.core'
]

MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
]

ROOT_URLCONF = 'toughradius.urls'

TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [os.path.join(PROJECT_ROOT, 'templates')],
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
            ],
        },
    },
]

WSGI_APPLICATION = 'toughradius.wsgi.application'


# Database
# https://docs.djangoproject.com/en/1.11/ref/settings/#databases

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.sqlite3',
        'NAME': os.path.join(PROJECT_ROOT, 'db.sqlite3'),
    }
}


# Password validation
# https://docs.djangoproject.com/en/1.11/ref/settings/#auth-password-validators

AUTH_PASSWORD_VALIDATORS = [
    {
        'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator',
    },
]


# Internationalization
# https://docs.djangoproject.com/en/1.11/topics/i18n/

# LANGUAGE_CODE = 'zh-CN'
LANGUAGE_CODE = 'en-US'

TIME_ZONE = 'UTC'

USE_I18N = True

USE_L10N = True

USE_TZ = True


# Static files (CSS, JavaScript, Images)
STATICFILES_DIRS = [os.path.join(PROJECT_ROOT, 'static')]
STATIC_URL = '/static/'

LOCALE_PATHS = [os.path.join(PROJECT_ROOT, 'locale')]

GUARDIAN_RAISE_403 = True
AUTHENTICATION_BACKENDS = (
    'django.contrib.auth.backends.ModelBackend',
    'guardian.backends.ObjectPermissionBackend',
)
GUARDIAN_GET_INIT_ANONYMOUS_USER = 'toughradius.apps.core.models.get_custom_anon_user'
PASSWORD_HASHERS = (
    'django.contrib.auth.hashers.MD5PasswordHasher',
    'django.contrib.auth.hashers.SHA1PasswordHasher',
)
AUTH_USER_MODEL = 'core.Operator'
