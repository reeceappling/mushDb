import {Note} from "@/app/components/formSubcomponents/notes";
import {AgarBatchInline, AssertAgarBatch, NewAgarBatchForm} from "@/app/components/agarBatchClient";
import RecentSelector, {SelectorProps} from "@/app/components/selector";
import {InlineProps} from "@/app/components/common";
import TestAndValidate from "@/app/components/testing/untested";
import {ACL} from "@/app/components/accessControlServer";

// TODO: CHANGE AGAR COLORS TO USE THEM FROM THE SERVER
export type AgarColor = "Clear" | "Black" | "Blue" | "Green" | "Yellow"| "Orange" | "Red";
export function TestAgarBatchOk() {
    const a: AgarBatchData = {
        _id: "(Batch ID HERE)",
        color: "clear",
        pcRun: "(Run ID HERE)",
        agarRecipe: "(Recipe ID HERE)",
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        // acl: {
        //     users: new Map<string, boolean>(), // TODO: FIXME!
        //     projects: new Map<string, boolean>(), // TODO: FIXME!
        //     blanketPerm: undefined,// TODO: FIXME!
        // }
    }
    return a
}

export interface AgarBatchData {
    _id: string
    color: string
    pcRun: string
    agarRecipe: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL // TODO: add everywhere necessary
}

export function AgarBatchSelector(sp: SelectorProps<AgarBatchData>) {
    // TODO: LOOK UP MOST RECENT AGAR BATCHES?
    // TODO: SELECT FROM THOSE BATCHES?
    // TODO: DONT USE RECENTSELECTOR!?
    return <TestAndValidate todos={["ensure agarBatch selector is working properly everywhere it is used"]}>
        <RecentSelector props={{
            allowCreation: sp.allowCreation,
            doSelect: sp.doSelect, // For selecting normally
            msgTxt: ChannelTextNewAgarBatch,
            recentEndpt: "agarBatches",
            assertType: AssertAgarBatch,
            closeTxt: "Close Batch List",
            createTxt: "Create Agar Batch",
            createEndpt: "agarBatch",
            lowercase: "agar batch",
            creatorInPage: sp.creatorInPage,
            inline: (inn: InlineProps<AgarBatchData>) => {
                return <AgarBatchInline data={inn.data} headerLevel={inn.headerLevel} onClick={inn.onClick}
                                        expandByDefault={inn.expandByDefault}/>
            },
            getId: (v: AgarBatchData) => {
                return v._id
            }
        }}>
            <NewAgarBatchForm handlers={{onCreate: sp.doSelect, isTopLevel: false/* TODO: ok?*/}}/>
        </RecentSelector></TestAndValidate>
}

export const ChannelTextNewAgarBatch = "newAgarBatch"