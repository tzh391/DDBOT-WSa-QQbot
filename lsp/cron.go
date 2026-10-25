package lsp

import (
	"fmt"
	"sync"

	"github.com/cnxysoft/DDBOT-WSa/lsp/cfg"
	"github.com/cnxysoft/DDBOT-WSa/lsp/mmsg"
	"github.com/cnxysoft/DDBOT-WSa/lsp/template"
	localutils "github.com/cnxysoft/DDBOT-WSa/utils"
	"github.com/sirupsen/logrus"
)

var cronLog = logrus.WithField("module", "cronjob")

type cronjobRun struct {
	*cfg.CronJob
	l *Lsp
}

func (c *cronjobRun) Run() {
	templateName := fmt.Sprintf("custom.cronjob.%s.tmpl", c.TemplateName)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for _, groupCode := range c.Target.Group {
			m, err := template.LoadAndExec(templateName, map[string]interface{}{
				"target": groupCode,
			})
			if err != nil {
				cronLog.WithError(err).WithField("template", templateName).
					Errorf("定时任务模板执行失败 (group=%d)", groupCode)
			}
			if m != nil {
				c.l.SendMsg(m, mmsg.NewGroupTarget(groupCode))
			}
		}
	}()
	go func() {
		defer wg.Done()
		for _, uin := range c.Target.Private {
			m, err := template.LoadAndExec(templateName, cronjobPrivateTemplateData(uin))
			if err != nil {
				cronLog.WithError(err).WithField("template", templateName).
					Errorf("定时任务模板执行失败 (private=%d)", uin)
			}
			if m != nil {
				c.l.SendMsg(m, mmsg.NewPrivateTarget(uin))
			}
		}
	}()
	wg.Wait()
}

// cronjobPrivateTemplateData 为私聊定时模板补充收件人的名称和 QQ 号。
func cronjobPrivateTemplateData(uin int64) map[string]interface{} {
	memberName := fmt.Sprintf("%d", uin)
	if b := localutils.GetBot(); b != nil {
		if friend := b.FindFriend(uin); friend != nil && friend.Nickname != "" {
			memberName = friend.Nickname
		}
	}
	return map[string]interface{}{
		"target":      uin,
		"member_code": uin,
		"member_name": memberName,
	}
}

func (l *Lsp) CronjobReload() {
	for _, entry := range l.cron.Entries() {
		l.cron.Remove(entry.ID)
	}
	cronjobs := cfg.GetCronJob()
	for _, entry := range cronjobs {
		if _, err := l.cron.AddJob(entry.Cron, &cronjobRun{entry, l}); err != nil {
			cronLog.WithField("cron_exp", entry.Cron).
				WithField("template_name", entry.TemplateName).
				WithField("target_group", entry.Target.Group).
				WithField("target_private", entry.Target.Private).
				Errorf("添加定时任务失败：%v", err)
		}
	}
}

func (l *Lsp) CronStart() {
	l.cron.Start()
}

func (l *Lsp) CronStop() {
	<-l.cron.Stop().Done()
}
