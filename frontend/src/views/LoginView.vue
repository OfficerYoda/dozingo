<template>
    <section class="login">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">Welcome Back</h2>
                        <small>Continue your bingo journey</small>
                    </div>
                    <form action="#">
                        <div class="mb-3">
                            <label for="username">Username or Email</label>
                            <div class="input-group">
                                <span><User :size="20" /></span>
                                <input type="text" id="username" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <div class="password-top">
                                <label for="password">Password</label>
                                <a href="#">Forgot Password?</a>
                            </div>
                            <div class="input-group">
                                <span><KeyRound :size="20" /></span>
                                <input type="password" id="password" required>
                            </div>
                        </div>
                        <button type="submit" class="btn btn-primary">Log In</button>
                    </form>
                    <small class="register-notice">
                        Don't have an account? <RouterLink to="/register">Join the Squad</RouterLink>
                    </small>
                </div>
            </div>
        </div>
    </section>
</template>

<style>
    .login .card{
        border-top: 5px solid var(--card-blue);
    }

    .heading{
        text-align: center;
    }

    .heading small{
        color: var(--color-text-subtle);
    }

    .input-group{
        display: flex;
        background-color: var(--color-bg-card-tinted);
        border-radius: var(--radius-sm);
        margin-top: 4px;
    }

    .input-group input{
        border-start-start-radius: 0;
        margin: 0;
    }

    .input-group span{
        display: flex;
        justify-content: center;
        align-items: center;
        padding-inline: 8px;
        color: var(--color-subheading);
    }

    .login input{
        background-color: var(--color-bg-card-tinted);
        outline: none;
        border: none;
        border-radius: var(--radius-sm);
        padding: 8px 10px;
        display: block;
        width: 100%;
    }

    .password-top{
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .password-top a{
        font-size: 0.7rem;
        font-weight: 700;
        color: var(--color-primary-600);
    }

    .login button{
        width: 100%;
        border-radius: 999px !important;
    }

    .register-notice{
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .register-notice a{
        font-weight: 600;
        color: var(--color-primary-600);
    }
</style>

<script setup lang="ts">
import { User, KeyRound } from 'lucide-vue-next'
</script>