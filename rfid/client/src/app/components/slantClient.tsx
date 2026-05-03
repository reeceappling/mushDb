'use client'

import React, {useState} from "react";
import NotesAreaOld, {IsValidNote, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    SlantData
} from "@/app/components/slantServer";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {AddToTransfers, InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayInput,
    DisposedSaleContamArea,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    NewEntryInput,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData, SendMultipartRequest, setFormData,
    setFormImages, SingleListProps,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    DisposedDisplay, ErrorDisplay,
    GensInlineDisplay, GensFormDisplay, MostRecentImageDisplay, OpenMainPage,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea, SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {AgarBatchArea, FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, InitialNotesState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {AgarBatchData, AgarBatchSelector} from "@/app/components/agarBatchServer";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {useCookies} from "react-cookie";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {OnViewCreatorsQuadColArea, OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {OvcForXfers} from "@/app/components/bagClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";

export function AssertSlant(input: any): asserts input is SlantData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Slant assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['agarBatch', 'string'],
        ['stickType','string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Slant assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function SlantImportDisplay({headerLevel, cookies}:ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const ImportSlant = () => {
        let formData = new FormData()
        if(species===undefined){
            setErr("Species must be set!")
            return
        }
        formData.set('data', JSON.stringify({
            created:created,
            stickType: stickType,
            species: species._id,
            subspecies: subspecies?._id,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
            //perms: perms, // TODO: validate on insert
        }))
        if(imageFile!==undefined){
            formData.set("image", imageFile, "imgFile")
        }


        SendMultipartRequest(BaseExternalUrl+"/db/import/slant", cookies, formData)
            .then(HandleTxtResponse)
            .then((newId) => {
                // TODO: maybe instead of redirecting, just give it a handler?
                redirect(BaseExternalUrl+"/view/slant/"+newId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
        {species?<ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel/*cookies={cookies}*/}/>:null}
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <GenerationArea readonly={false} updateParent={setGeneration} headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImageFile}/>
        <SlantStickSelector setStickType={setStickType}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} headerLevel={headerLevel}/>
        <button className={"greenButton"} onClick={ImportSlant}>{"Import Slant"}</button>
    </ImportEntryFormWrapper>
}

export default function SlantDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertSlant(data)
        const [initial, setInitial] = useState(data)
        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: SlantData)=>{
            setInitial(updated)
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setAcl(updated.acl)
        }
        const slantSubmit = ()=>{
            let body = new FormData()
            let dataObj:any={
                knownFruitable:knownFruitable,
                sale: sale,
                disposed:disposed,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(images)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl+"/db/update/slant/"+initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSlant(entry)
                    updateInitial(entry)
                    //window.location.reload()
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: anything here?
        ] // TODO: THIS!
        return (
            <DisplayFormWrapper entryType={"slant"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Slant"} entryType={"slant"} />
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed}
                                                    readonly={readonly} setDisposedOnParent={setDisposed}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                        <AgarBatchArea agarBatchId={initial.agarBatch} headerLevel={headerLevel}/>
                        <div>
                            {"Stick type:"+(initial.stickType || "none")}
                        </div>
                    </FlexedSinglesGroup>
                    {/* TODO: CondensationCoverageAtSealTimeField `bson:"inline"` // Percentage of condensation surface area coverage at seal time
                PourCoverageField                   `bson:"inline"` // Percentage of bottom surface area agar coverage
                WetAtCooledTimeField                `bson:"inline"` // Wet when initially cooled? True, false, or unknown
                AgarOnOutsideAtPourTimeFiel*/}
                    <FlexedSinglesGroup>
                        <InnocDisplay innoc={initial.innoc}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                                  canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedArea>

                <TransfersOutDisplay thisId={initial._id} thisEntryType={"slant"} transfersOut={transfersOut} allowNewTransferCreation={!readonly} cookies={cookies}/>
                <PicsDisplay pix={images} updateParent={setImages} readonly={readonly} headerLevel={headerLevel}/>{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                {readonly ? null : <input type="submit" value="Update" onClick={slantSubmit} onSubmit={(e)=>{e.preventDefault();}}/>}
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Slant data format incorrect: " + err}</div>
    }
}

export function NewSlantForm({handlers,agarBatchIn}: {handlers: NewEntryInput<SlantData>, agarBatchIn?:AgarBatchData}){
    const [agarBatch, setAgarBatch] = useState(agarBatchIn)
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle handlers.isTopLevel
    const createSlant = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(agarBatch===undefined){
            setErr("An agar batch must be selected")
            return
        }
        let body: any = {
            agarBatch: agarBatch,
            stickType: stickType,
            writeTagTo: writeTagTo,
        }
        fetch(BaseExternalUrl+"/create/slant", {
            method: "POST",
            headers: {
                credentials: 'include',
                // TODO: may need 'Cookie': cookies,
                'Content-type': "application/json"
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertSlant(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err}/>
        <SlantStickSelector setStickType={setStickType}/>
        <AgarBatchSelector doSelect={setAgarBatch} allowCreation={true} creatorInPage={true}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createSlant} onSubmit={(e)=>{e.preventDefault();}}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SlantStickSelector(x: { setStickType: (s?: string) => void }){
    return <div>
        {"Slant stick: "}<select className={"tailwindSelector"} defaultValue={"none"} disabled={false} onChange={(e) => {x.setStickType((e.currentTarget.value==="none")?undefined:e.currentTarget.value)}}>
        {["none","tongueDepressor"].map((s, i: number) => {
            return <option value={s} key={i}>{s}</option>
        })}
    </select>
    </div>
}

export function SlantInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<SlantData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Slant"} entryType={"slant"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <SpeciesArea readonly={true} initial={data.species}  />
            <SubspeciesArea readonly={true} currentSpecies={data.species} initialSub={data.subspecies} />
            <KnownFruitableArea initial={data.knownFruitable} readonly={true} />
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore} />
            <DisposedSaleContamArea sale={data.sale} disposed={data.disposed} contams={data.contamination} />
        </InlineSubArea>
        <InlineExpansionArea props={{expanded:expanded}}>
            <AgarBatchArea agarBatchId={data.agarBatch} offset={-1}/>
            {/* TODO: SLANT STICK */}
            {/*TODO: <ProjectsArea allowCreate={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel} offset={-1} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function SlantListDisplay({data, onClick}: SingleListProps<SlantData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <SlantInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }