#pragma once

#include <kstd/memory/UniquePtr.hh>
#include <kstd/async/Core.hh>

#include "Core.hh"
#include "actions/Action.hh"

namespace juno {

class Job {
public:
    enum class Type { oneshot, repeated, rule };

    explicit Job(kstd::UniquePtr<Action> action) : m_action(std::move(action)) {}

    kstd::Coro<bool> process() {
        if (co_await isReady()) co_await m_action->execute();
        co_return co_await isDone();
    }

private:
    virtual kstd::Coro<bool> isReady() = 0;
    virtual kstd::Coro<bool> isDone()  = 0;

    kstd::UniquePtr<Action> m_action;
};

class OneShotJob : public Job {
public:
private:
};

class RepeatedJob : public Job {
public:
private:
};

class RuleJob : public Job {
public:
private:
};

}  // namespace juno
