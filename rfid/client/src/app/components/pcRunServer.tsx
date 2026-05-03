import {Note} from "@/app/components/formSubcomponents/notes";
import {AssertPcRun, NewPcRunForm, PcRunInline} from "@/app/components/pcRunClient";
import RecentSelector, {SelectorProps} from "@/app/components/selector";
import {InlineProps} from "@/app/components/common";
import {ACL} from "@/app/components/accessControlServer";

export function TestPcRunOk(){
    const a: PcRunData = {
        _id: "(ID_HERE)",
        creationDate: Date.now()-2000,
        runtimeMinutes: 120,
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        },{
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: Date.now(),
    }
    return a
}

export interface PcRunData {
    _id: string
    creationDate: number
    runtimeMinutes: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function RecentPCRunSelector(sp: SelectorProps<PcRunData>){ // TODO: likely overhaul
    return <RecentSelector props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally
        msgTxt: ChannelTextNewPcRun,
        recentEndpt: "pcRuns",
        assertType: AssertPcRun,
        closeTxt: "Close Run List",
        createTxt: "Create Pc Run",
        createEndpt: "pcRun",
        lowercase: "pc run",
        creatorInPage: sp.creatorInPage,
        inline: (inn: InlineProps<PcRunData>)=>{
            return <PcRunInline data={inn.data} headerLevel={inn.headerLevel} onClick={inn.onClick} expandByDefault={inn.expandByDefault}/>
        },
        getId: (v: PcRunData)=>{
            return v._id
        }
    }}>
        <NewPcRunForm handlers={{onCreate:sp.doSelect, isTopLevel: false}} />
    </RecentSelector>
    // RecentSelector<PcRunData>({
    //     msgTxt: ChannelTextNewPcRun,
    //     recentEndpt: "pcRuns",
    //     assertType: AssertPcRun,
    //     closeTxt: "Close PC Runs List",
    //     createTxt: "Create PC Run",
    //     //newForm: NewPcRunForm, // TODO: REENABLE! NEED!
    //     createEndpt: "pcRun",
    //     lowercase: "pc run",
    //     inline: (inlineIn: InlineProps<PcRunData>)=>{return PcRunInline(inlineIn)},
    // })(sp)
}

const ChannelTextNewPcRun = "newPcRun"