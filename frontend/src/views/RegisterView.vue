<template>
    <section class="register">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">Join the Game</h2>
                        <small>Start your journey here at dozingo</small>
                    </div>
                    <form action="#">
                        <div class="mb-3">
                            <label for="username">Username</label>
                            <div class="input-group">
                                <span><User :size="20" /></span>
                                <input type="text" id="username" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="email">Email</label>
                            <div class="input-group">
                                <span><Mail :size="20" /></span>
                                <input type="email" id="email" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="password">Password</label>
                            <div class="input-group">
                                <span><KeyRound :size="20" /></span>
                                <input type="password" id="password" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="confirm-password">Confirm Password</label>
                            <div class="input-group">
                                <span><KeyRound :size="20" /></span>
                                <input type="password" id="confirm-password" required>
                            </div>
                        </div>
                        <button type="submit" class="btn btn-primary">Log In</button>
                    </form>
                    <small class="login-notice">
                        Already have an account? <RouterLink to="/login">Log In</RouterLink>
                    </small>
                </div>
            </div>
        </div>
    </section>
</template>

<style>
    .register .card{
        border-top: 5px solid var(--card-blue);
    }

    .heading{
        text-align: center;
    }

    .heading small{
        color: var(--color-text-subtle);
    }

    .register input{
        background-color: var(--color-bg-card-tinted);
        outline: none;
        border: none;
        border-radius: var(--radius-sm);
        padding: 8px 10px;
        display: block;
        width: 100%;
    }

    .register button{
        width: 100%;
        border-radius: 999px !important;
    }

    .login-notice{
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .login-notice a{
        font-weight: 600;
        color: var(--color-primary-600);
    }
</style>

<script setup lang="ts">
import { User, Mail, KeyRound } from 'lucide-vue-next'
</script>